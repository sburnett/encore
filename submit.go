package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/sburnett/encore/store"
)

type submitState struct {
	results chan *store.Result
}

var submissionCount = metrics.GetOrRegisterCounter("ResultsSubmitted", nil)
var submissionErrorCount = metrics.GetOrRegisterCounter("ResultSubmissionRequestsMalformed", nil)

func NewSubmissionServer(s store.Store) *submitState {
	resultsChan := make(chan *store.Result)
	go s.WriteResults(resultsChan)

	return &submitState{
		results: resultsChan,
	}
}

func (state *submitState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	submissionCount.Inc(1)

	// Let clients post results from any domain. This is necessary because
	// our measurements run and report from third party Web sites.
	w.Header().Add("Access-Control-Allow-Origin", "*")

	var rawRequest bytes.Buffer
	if err := r.Write(&rawRequest); err != nil {
		log.Print("error writing HTTP request")
		w.WriteHeader(http.StatusInternalServerError)
		submissionErrorCount.Inc(1)
		return
	}
	log.Printf("inserting new result from '%v'", r.RemoteAddr)

	w.WriteHeader(http.StatusOK)

	state.results <- &store.Result{
		Timestamp:  time.Now(),
		RemoteAddr: r.RemoteAddr,
		RawRequest: rawRequest.Bytes(),
	}
}
