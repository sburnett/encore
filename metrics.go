package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/influxdb"
)

var influxDbHost, influxDbDatabase, influxDbUsername, influxDbPassword string
var influxDbExportInterval time.Duration

func init() {
	flag.StringVar(&influxDbHost, "influx_db_host", "127.0.0.1:8086", "InfluxDB host")
	flag.StringVar(&influxDbDatabase, "influx_db_database", "encore", "InfluxDB database")
	flag.StringVar(&influxDbUsername, "influx_db_username", "", "InfluxDB username")
	flag.StringVar(&influxDbPassword, "influx_db_password", "", "InfluxDB password")
	flag.DurationVar(&influxDbExportInterval, "influx_db_export_interval", time.Minute, "Export stats to InfluxDB once every interval")
}

func initMetrics() {
	go influxdb.Influxdb(metrics.DefaultRegistry, influxDbExportInterval, &influxdb.Config{
		Host:     influxDbHost,
		Database: influxDbDatabase,
		Username: influxDbUsername,
		Password: influxDbPassword,
	})
	go metrics.Log(metrics.DefaultRegistry, 1e9, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
}
