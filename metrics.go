package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	//"github.com/rcrowley/go-metrics/influxdb"
	"github.com/vrischmann/go-metrics-influxdb"
)

var influxDbHost, influxDbDatabase, influxDbUsername, influxDbPassword string
var influxDbExportInterval time.Duration
var printMetrics bool

func init() {
	flag.StringVar(&influxDbHost, "influx_db_host", "127.0.0.1:8086", "InfluxDB host")
	flag.StringVar(&influxDbDatabase, "influx_db_database", "encore", "InfluxDB database")
	flag.StringVar(&influxDbUsername, "influx_db_username", "", "InfluxDB username")
	flag.StringVar(&influxDbPassword, "influx_db_password", "", "InfluxDB password")
	flag.DurationVar(&influxDbExportInterval, "influx_db_export_interval", time.Minute, "Export stats to InfluxDB once every interval")
	flag.BoolVar(&printMetrics, "print_metrics", false, "Print all metrics to stderr")
}

func initMetrics() {
	go influxdb.InfluxDB(metrics.DefaultRegistry, influxDbExportInterval, 
		influxDbHost,
		influxDbDatabase,
		influxDbUsername,
		influxDbPassword,
	)
	if printMetrics {
		go metrics.Log(metrics.DefaultRegistry, 1e9, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}
}
