package store

import (
	"database/sql"
	"flag"
	"log"
	"net"
	"time"
)

type Schedule struct {
	Id               int
	Priority         int
	MaxDuration      time.Duration
	MaxMeasurements  sql.NullInt64
	MaxRatePerSecond sql.NullInt64
	Tasks            []Task
}

type Task struct {
	Id                 int
	TemplateParameters map[string]sql.NullString
}

type Query struct {
	Id             int
	Timestamp      time.Time
	RemoteAddr     string
	RawRequest     []byte
	Task           int
	Substrate      string
	ParametersJson []byte
	ResponseBody   []byte
}

type ParsedQuery struct {
	Query          int
	MeasurementId  string
	Timestamp      time.Time
	ClientIp       net.IP
	ClientLocation string
	Substrate      string
	Parameters     map[string]sql.NullString
}

type Result struct {
	Id         int
	Timestamp  time.Time
	RemoteAddr string
	RawRequest []byte
}

type ParsedResult struct {
	Result         int
	Timestamp      time.Time
	MeasurementId  string
	Outcome        string
	Origin         string
	Referer        string
	ClientIp       net.IP
	ClientLocation string
	UserAgent      string
}

type CountResultsRequest struct {
	Referer  string
	Response chan CountResultsResponse
}

type CountResultsResponse struct {
	Count int
	Err   error
}

type ResultsPerDayRequest struct {
	Referer  string
	Response chan ResultsPerDayResponse
}

type ResultsPerDayResponse struct {
	Results map[string]int
	Err     error
}

type ResultsPerCountryRequest struct {
	Referer  string
	Response chan ResultsPerCountryResponse
}

type ResultsPerCountryResponse struct {
	Results map[string]int
	Err     error
}

type Store interface {
	Close()
	Schedules() <-chan *Schedule
	WriteTasks(tasks <-chan *Task)
	WriteQueries(queries <-chan *Query)
	Queries() <-chan *Query
	UnparsedQueries() <-chan *Query
	WriteParsedQueries(queries <-chan *ParsedQuery)
	WriteResults(results <-chan *Result)
	Results() <-chan *Result
	UnparsedResults() <-chan *Result
	WriteParsedResults(results <-chan *ParsedResult)
	CountResultsForReferrer(requests <-chan CountResultsRequest)
	ResultsPerDayForReferrer(requests <-chan ResultsPerDayRequest)
	ResultsPerCountryForReferrer(requests <-chan ResultsPerCountryRequest)
	ComputeResultsTables() error
}

var databaseDriver, databaseName string

func init() {
	flag.StringVar(&databaseDriver, "driver", "postgres", "Database driver")
	flag.StringVar(&databaseName, "database", "dbname=encore sslmode=disable", "Name or path of the database to use")
}

func Open() Store {
	db, err := sql.Open(databaseDriver, databaseName)
	if err != nil {
		log.Fatalf("error opening database %s: %v", databaseName, err)
	}
	switch databaseDriver {
	case "postgres":
		return openPostgres(db)
	default:
		log.Fatalf("invalid database driver")
	}
	return nil
}
