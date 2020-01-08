package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var opts struct {
	Listen        string `short:"l" long:"listen" description:"Listen address" value-name:"[HOST]:PORT" default:":9550"`
	MetricsPath   string `short:"m" long:"metrics-path" description:"Metrics path" value-name:"PATH" default:"/scrape"`
	V2rayEndpoint string `short:"e" long:"v2ray-endpoint" description:"V2Ray API endpoint" value-name:"HOST:PORT" default:"127.0.0.1:8080"`
	Version       bool   `long:"version" description:"Show version"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var exporter *Exporter

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(
		exporter.registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
	).ServeHTTP(w, r)
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Printf("v2ray-exporter %v (commit %v, built %v)\n", buildVersion, buildCommit, buildDate)
		os.Exit(0)
	}

	exporter = NewExporter(opts.V2rayEndpoint)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/scrape", scrapeHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>V2Ray Exporter ` + buildVersion + `</title></head>
<body>
<h1>V2Ray Exporter ` + buildVersion + `</h1>
<p><a href='` + opts.MetricsPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
