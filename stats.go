package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/sburnett/encore/store"
)

type statsState struct {
	CountResultsRequests      chan store.CountResultsRequest
	ResultsPerDayRequests     chan store.ResultsPerDayRequest
	ResultsPerCountryRequests chan store.ResultsPerCountryRequest
}

var refererRedirects = expvar.NewInt("StatsRefererRedirects")
var statsHits = expvar.NewInt("StatsHits")
var statsTemplateExecutionErrorCount = expvar.NewInt("StatsTemplateExecutionError")

func NewStatsServer(s store.Store, templatesPath string) http.Handler {
	countResultsRequests := make(chan store.CountResultsRequest)
	go s.CountResultsForReferrer(countResultsRequests)

	resultsPerDayRequests := make(chan store.ResultsPerDayRequest)
	go s.ResultsPerDayForReferrer(resultsPerDayRequests)

	resultsPerCountryRequests := make(chan store.ResultsPerCountryRequest)
	go s.ResultsPerCountryForReferrer(resultsPerCountryRequests)

	return &statsState{
		CountResultsRequests:      countResultsRequests,
		ResultsPerDayRequests:     resultsPerDayRequests,
		ResultsPerCountryRequests: resultsPerCountryRequests,
	}
}

func refererRedirect(w http.ResponseWriter, r *http.Request) {
	refererRedirects.Add(1)

	var referer string
	referers, ok := r.Header["Referer"]
	if !ok {
		referer = ""
	} else {
		referer = referers[0]
	}

	parameters := url.Values{}
	parameters.Set("referer", referer)
	parameters.Encode()

	redirectUrl := url.URL{
		Path:     "/stats.html",
		RawQuery: parameters.Encode(),
	}
	log.Print(redirectUrl.String())
	http.Redirect(w, r, redirectUrl.String(), http.StatusFound)
}

func formatReferer(refererString string) (string, error) {
	referer, err := url.ParseRequestURI(refererString)
	if err != nil {
		invalidRefererCount.Add(1)
		return "", fmt.Errorf("invalid referer")
	}
	referer.RawQuery = "" // Remove query parameters for robustness.

	return referer.String(), nil
}

func countResults(requests chan store.CountResultsRequest, referer string) (int, error) {
	request := store.CountResultsRequest{
		Referer:  referer,
		Response: make(chan store.CountResultsResponse),
	}
	requests <- request
	response := <-request.Response
	return response.Count, response.Err
}

func resultsPerDay(requests chan store.ResultsPerDayRequest, referer string) (map[string]int, error) {
	request := store.ResultsPerDayRequest{
		Referer:  referer,
		Response: make(chan store.ResultsPerDayResponse),
	}
	requests <- request
	response := <-request.Response
	return response.Results, response.Err
}

func resultsPerCountry(requests chan store.ResultsPerCountryRequest, referer string) (map[string]int, error) {
	request := store.ResultsPerCountryRequest{
		Referer:  referer,
		Response: make(chan store.ResultsPerCountryResponse),
	}
	requests <- request
	response := <-request.Response
	return response.Results, response.Err
}

func (state *statsState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statsHits.Add(1)

	referer := r.URL.Query().Get("referer")
	refererString, err := formatReferer(referer)
	if err != nil {
		log.Printf("error formatting referer: %v", err)
		refererString = ""
	}

	totalResults, err := countResults(state.CountResultsRequests, refererString)
	if err != nil {
		log.Printf("error counting results for this referer: %s", err)
		totalResults = 0
	}

	perDay, err := resultsPerDay(state.ResultsPerDayRequests, refererString)
	if err != nil {
		log.Printf("error counting results per day for this referer: %s", err)
		perDay = map[string]int{}
	}

	perCountry, err := resultsPerCountry(state.ResultsPerCountryRequests, refererString)
	if err != nil {
		log.Printf("error counting results per country for this referer: %s", err)
		perCountry = map[string]int{}
	}

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(struct {
		Site              string
		TotalResults      int
		ResultsPerDay     map[string]int
		ResultsPerCountry map[string]int
	}{
		Site:              referer,
		TotalResults:      totalResults,
		ResultsPerDay:     perDay,
		ResultsPerCountry: perCountry,
	}); err != nil {
		log.Printf("error encoding JSON result: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
