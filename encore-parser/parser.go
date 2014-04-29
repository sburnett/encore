package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/abh/geoip"
	"github.com/sburnett/encore/store"
)

func parseQueries(queries <-chan *store.Query, geolocator *geoip.GeoIP) <-chan *store.ParsedQuery {
	parsedQueries := make(chan *store.ParsedQuery)
	go func() {
		for query := range queries {
			request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(query.RawRequest)))
			if err != nil {
				log.Printf("error parsing result request: %v", err)
			}
			clientIp := request.Header.Get("X-Real-Ip")
			if clientIp == "" {
				clientIp = query.RemoteAddr
			}

			var parameters map[string]string
			if err := json.Unmarshal(query.ParametersJson, &parameters); err != nil {
				log.Printf("error parsing query parameters: %v", err)
			}

			parametersNullable := make(map[string]sql.NullString)
			for k, v := range parameters {
				parametersNullable[k] = sql.NullString{
					String: v,
					Valid:  true,
				}
			}

			host, _, err := net.SplitHostPort(clientIp)
			if err != nil {
				host = clientIp
			}

			country, _ := geolocator.GetCountry(host)

			parsedQueries <- &store.ParsedQuery{
				Query:          query.Id,
				MeasurementId:  parameters["measurementId"],
				Timestamp:      query.Timestamp,
				ClientIp:       net.ParseIP(host),
				ClientLocation: country,
				Substrate:      query.Substrate,
				Parameters:     parametersNullable,
			}
		}
		close(parsedQueries)
	}()
	return parsedQueries
}

func parseResults(results <-chan *store.Result, geolocator *geoip.GeoIP) <-chan *store.ParsedResult {
	parsedResults := make(chan *store.ParsedResult)
	go func() {
		for result := range results {
			request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(result.RawRequest)))
			if err != nil {
				log.Printf("error parsing result request: %v", err)
			}
			measurementId := request.URL.Query().Get("cmh-id")
			outcome := request.URL.Query().Get("cmh-result")
			message := request.URL.Query().Get("cmh-message")
			userAgent := request.Header.Get("User-Agent")
			origin := request.Header.Get("Origin")
			referer := request.Header.Get("Referer")
			clientIp := request.Header.Get("X-Real-Ip")
			if clientIp == "" {
				clientIp = result.RemoteAddr
			}

			host, _, err := net.SplitHostPort(clientIp)
			if err != nil {
				host = clientIp
			}

			country, _ := geolocator.GetCountry(host)

			parsedResults <- &store.ParsedResult{
				Result:         result.Id,
				Timestamp:      result.Timestamp,
				MeasurementId:  measurementId,
				Outcome:        outcome,
				Message:        message,
				Origin:         origin,
				Referer:        referer,
				ClientIp:       net.ParseIP(host),
				ClientLocation: country,
				UserAgent:      userAgent,
			}
		}
		close(parsedResults)
	}()
	return parsedResults

}

func main() {
	var geoipDatabase string
	flag.StringVar(&geoipDatabase, "geoip_database", "/usr/share/GeoIP/GeoIP.dat", "Path of GeoIP database")
	flag.Parse()

	log.Printf("starting")

	s := store.Open()
	defer s.Close()

	geolocator, err := geoip.Open(geoipDatabase)
	if err != nil {
		panic(err)
	}

	queries := s.UnparsedQueries()
	parsedQueries := parseQueries(queries, geolocator)
	s.WriteParsedQueries(parsedQueries)

	results := s.UnparsedResults()
	parsedResults := parseResults(results, geolocator)
	s.WriteParsedResults(parsedResults)

	if err := s.ComputeResultsTables(); err != nil {
		panic(err)
	}

	log.Printf("done")
}
