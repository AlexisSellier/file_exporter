package main

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"log"
	"path/filepath"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress		= flag.String("web.listen-address", ":9121", "Address to listen on for web interface and telemetry.")
	metricPath		= flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	configFile       	= flag.String("config", "./config.json", "config file location.")
)

type exporter struct {
	files  []string
	lastModifiedTime *prometheus.Desc
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.lastModifiedTime
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	for _, file := range e.files {
		var value int64 = 0
		fileinfo, err := os.Stat(file)
		if err == nil {
			value = fileinfo.ModTime().Unix()
		}
		ch <- prometheus.MustNewConstMetric(e.lastModifiedTime, prometheus.GaugeValue, float64(value), filepath.Clean(file))
	}
}

func newExporter(configfile string) *exporter {
	var e exporter
	e.lastModifiedTime = prometheus.NewDesc("file_last_modified_time", "Last Modified Time", []string{"filename"}, nil)
	file, err := ioutil.ReadFile(configfile)
	if err != nil {
		fmt.Printf("Cannot open config file: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(file, &e.files)
	return &e
}

func main() {
	flag.Parse()
	e := newExporter(*configFile)
	prometheus.MustRegister(e)
	prometheus.Unregister(prometheus.NewGoCollector())
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	http.Handle(*metricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		       <head><title>Last modified file  exporter</title></head>
		       <body>
		       <h1>Burrow exporter</h1>
		       <p><a href='` + *metricPath + `'>Metrics</a></p>
		       </body>
		       </html>`))
	})
	log.Printf("providing metrics at %s%s", *listenAddress, *metricPath)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

