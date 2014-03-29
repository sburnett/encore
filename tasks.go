package main

import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/maxhauser/jsmin"
	"github.com/sburnett/encore/store"
)

type measurementsServerState struct {
	Templates            *template.Template
	Queries              chan *store.Query
	Store                store.Store
	TaskGroups           <-chan []store.Task
	MeasurementIds       <-chan string
	CountResultsRequests chan store.CountResultsRequest
	ServerUrl            string
}

const hintPrefix string = "cmh-"

const (
	JavaScriptExtension string = ".js"
	HtmlExtension              = ".html"
)

var requestCount = expvar.NewInt("TasksRequested")
var optOutCount = expvar.NewInt("OptOut")
var noViableTasksCount = expvar.NewInt("NoViableTask")
var countResultsErrorCount = expvar.NewInt("CountResultsError")
var templateExecutionErrorCount = expvar.NewInt("TemplateExecutionError")
var requestParseErrorCount = expvar.NewInt("RequestParseError")
var parametersMarshalErrorCount = expvar.NewInt("ParametersMarshalError")
var minifiedCount = expvar.NewInt("Minified")
var unminifiedCount = expvar.NewInt("Unminified")
var responseCount = expvar.NewInt("TasksServed")
var noRefererCount = expvar.NewInt("NoReferer")
var invalidRefererCount = expvar.NewInt("InvalidReferer")
var taskGroupTimeoutCount = expvar.NewInt("TaskGroupTimeout")
var missingTaskTypeCount = expvar.NewInt("MissingTaskType")

func NewTaskServer(s store.Store, serverUrl, templatesPath string) *measurementsServerState {
	queries := make(chan *store.Query)
	go s.WriteQueries(queries)

	measurementIds := generateMeasurementIds()

	go s.ScheduleTaskGroups()

	taskGroups := s.TaskGroups()

	countResultsRequests := make(chan store.CountResultsRequest)
	go s.CountResultsForReferrer(countResultsRequests)

	return &measurementsServerState{
		Store:                s,
		Templates:            template.Must(template.ParseGlob(filepath.Join(templatesPath, "[a-zA-Z]*"))),
		Queries:              queries,
		MeasurementIds:       measurementIds,
		TaskGroups:           taskGroups,
		CountResultsRequests: countResultsRequests,
		ServerUrl:            serverUrl,
	}
}

func parseContentType(path string) string {
	switch filepath.Ext(path) {
	case ".js":
		return JavaScriptExtension
	case ".html", ".htm":
		return HtmlExtension
	default:
		return HtmlExtension
	}
}

func parseHints(r *http.Request) (hints map[string]string) {
	hints = make(map[string]string)

	for key, values := range r.URL.Query() {
		hints[key] = values[0]
	}

	for _, cookie := range r.Cookies() {
		if cookie == nil {
			continue
		}
		if !strings.HasPrefix(cookie.Name, hintPrefix) {
			continue
		}
		k := strings.TrimPrefix(cookie.Name, hintPrefix)
		hints[k] = cookie.Value
	}

	if !debugMode {
		return
	}

	referers, ok := r.Header["Referer"]
	if !ok {
		return
	}
	referer, err := url.ParseRequestURI(referers[0])
	if err != nil {
		return
	}
	queries := referer.Query()
	for key, values := range queries {
		if !strings.HasPrefix(key, hintPrefix) {
			continue
		}
		k := strings.TrimPrefix(key, hintPrefix)
		v := values[0]
		hints[k] = v
	}

	return
}

func countResultsForReferer(requests chan store.CountResultsRequest, r *http.Request) (int, error) {
	referers, ok := r.Header["Referer"]
	if !ok {
		noRefererCount.Add(1)
		return 0, fmt.Errorf("no referer")
	}
	referer, err := url.ParseRequestURI(referers[0])
	if err != nil {
		invalidRefererCount.Add(1)
		return 0, fmt.Errorf("invalid referer")
	}
	referer.RawQuery = "" // Remove query parameters for robustness.

	return countResults(requests, referer.String())
}

func (state *measurementsServerState) selectTask(hints map[string]string) *store.Task {
	var taskGroup []store.Task
	select {
	case taskGroup = <-state.TaskGroups:
	case <-time.After(time.Second):
		taskGroupTimeoutCount.Add(1)
	}

	if taskGroup == nil || len(taskGroup) == 0 {
		return nil
	}

	return &taskGroup[rand.Intn(len(taskGroup))]
}

func (state *measurementsServerState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestCount.Add(1)
	log.Printf("serving %v", r.URL)

	substrate := parseContentType(r.URL.Path)
	switch substrate {
	case HtmlExtension:
		w.Header().Set("Content-Type", "text/html")
	case JavaScriptExtension:
		w.Header().Set("Content-Type", "application/javascript")
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	hints := parseHints(r)

	if disabled, ok := hints["disable"]; ok && disabled == "true" {
		log.Printf("user opted out of Encore")
		w.WriteHeader(http.StatusOK)
		optOutCount.Add(1)
		return
	}

	// Select a task template
	task := state.selectTask(hints)
	if task == nil {
		log.Printf("cannot find viable task")
		w.WriteHeader(http.StatusInternalServerError)
		noViableTasksCount.Add(1)
		return
	}

	taskType, ok := task.Parameters["taskType"]
	if !ok || !taskType.Valid {
		log.Printf("malformed task: missing taskType")
		w.WriteHeader(http.StatusInternalServerError)
		missingTaskTypeCount.Add(1)
		return
	}

	templateName := taskType.String + substrate

	// Select task parameters
	taskParameters := make(map[string]string)
	taskParameters["serverUrl"] = state.ServerUrl
	taskParameters["measurementId"] = <-state.MeasurementIds
	taskParameters["hintJQueryAlreadyLoaded"] = hints["jQueryAlreadyLoaded"]
	taskParameters["hintShowStats"] = hints["showStats"]
	if showStats, ok := hints["showStats"]; !ok || showStats != "false" {
		count, err := countResultsForReferer(state.CountResultsRequests, r)
		if err != nil {
			log.Printf("error counting results: %v", err)
			countResultsErrorCount.Add(1)
		}
		taskParameters["count"] = fmt.Sprint(count)
	}
	for k, v := range task.Parameters {
		if !v.Valid {
			continue
		}
		taskParameters[k] = v.String
	}

	// Execute the template
	responseBody := bytes.Buffer{}
	if err := state.Templates.ExecuteTemplate(&responseBody, templateName, taskParameters); err != nil {
		log.Printf("error executing task template %s: %v", templateName, err)
		w.WriteHeader(http.StatusInternalServerError)
		templateExecutionErrorCount.Add(1)
		return
	}

	if minify, ok := hints["minify"]; ok && minify == "false" {
		responseBody.WriteTo(w)
		minifiedCount.Add(1)
	} else {
		jsmin.Run(&responseBody, w)
		unminifiedCount.Add(1)
	}

	var rawRequest bytes.Buffer
	if err := r.Write(&rawRequest); err != nil {
		log.Print("error writing HTTP request")
		w.WriteHeader(http.StatusInternalServerError)
		requestParseErrorCount.Add(1)
		return
	}
	parametersBytes, err := json.Marshal(taskParameters)
	if err != nil {
		log.Printf("cannot marshal task parameters to JSON")
		w.WriteHeader(http.StatusInternalServerError)
		parametersMarshalErrorCount.Add(1)
		return
	}

	state.Queries <- &store.Query{
		Timestamp:      time.Now(),
		RemoteAddr:     r.RemoteAddr,
		RawRequest:     rawRequest.Bytes(),
		Task:           task.Id,
		Substrate:      substrate,
		ParametersJson: parametersBytes,
		ResponseBody:   responseBody.Bytes(),
	}

	responseCount.Add(1)
}
