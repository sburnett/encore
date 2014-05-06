package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ParsePlatform/go.grace/gracehttp"
	"github.com/sburnett/cube"
	"github.com/sburnett/encore/store"
)

var debugMode bool

func main() {
	var listenAddress, serverUrl, taskTemplatesPath, statsTemplatesPath, staticPath, cubeCollectionType, logfile, geoipDatabase string
	flag.BoolVar(&debugMode, "debug", false, "Enable parsing of cmh- debug parameters in requests")
	flag.StringVar(&listenAddress, "listen_address", "127.0.0.1:8080", "")
	flag.StringVar(&serverUrl, "server_url", "http://localhost:8080", "URL that clients should use to contact this server.")
	flag.StringVar(&taskTemplatesPath, "task_templates_path", "task-templates", "Path to task templates")
	flag.StringVar(&statsTemplatesPath, "stats_templates_path", "stats-templates", "Path to stats templates")
	flag.StringVar(&staticPath, "static_path", "static", "Path to static content to serve")
	flag.StringVar(&cubeCollectionType, "cube_collection_type", "encore", "Use this label for statistics we send to Cube")
	flag.StringVar(&logfile, "logfile", "", "Write logs to this file instead of stdout")
	flag.StringVar(&geoipDatabase, "geoip_database", "/usr/share/GeoIP/GeoIP.dat", "Path of GeoIP database")
	flag.Parse()

	printVersionIfAsked()

	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("error opening logfile: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	log.Printf("starting")

	initMetrics()

	go cube.Run(cubeCollectionType)

	s := store.Open()
	defer s.Close()

	tasksServer := NewTaskServer(s, serverUrl, taskTemplatesPath, geoipDatabase)
	submissionServer := NewSubmissionServer(s)
	statsServer := NewStatsServer(s, statsTemplatesPath)

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(staticPath))))
	mux.Handle("/task.js", tasksServer)
	mux.Handle("/task.html", tasksServer)
	mux.Handle("/submit", submissionServer)
	mux.HandleFunc("/version", versionServer)
	mux.Handle("/stats/", statsServer)
	mux.HandleFunc("/stats/refer", refererRedirect)
	server := http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}

	log.Printf("serving at %s", listenAddress)
	if err := gracehttp.Serve(&server); err != nil {
		panic(err)
	}

	log.Printf("exiting")
}
