package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/rcrowley/go-metrics"
)

func generateMeasurementIds() <-chan string {
	measurementIds := make(chan string)
	measurementIdCounter := metrics.GetOrRegisterCounter("MeasurementIdsGenerated", nil)
	go func() {
		r := rand.NewSource(time.Now().UnixNano())
		for {
			id := fmt.Sprintf("%016x", r.Int63())
			measurementIds <- id
			measurementIdCounter.Inc(1)
		}
	}()
	return measurementIds
}
