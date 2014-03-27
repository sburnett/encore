package main

import (
	"expvar"
	"fmt"
	"math/rand"
	"time"
)

func generateMeasurementIds() <-chan string {
	measurementIds := make(chan string)
	measurementIdCounter := expvar.NewInt("MeasurementIdsGenerated")
	go func() {
		r := rand.NewSource(time.Now().UnixNano())
		for {
			id := fmt.Sprintf("%016x", r.Int63())
			measurementIds <- id
			measurementIdCounter.Add(1)
		}
	}()
	return measurementIds
}
