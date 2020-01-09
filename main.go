package main

import (
	"fmt"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	Listen        string `short:"l" long:"listen" description:"Listen address" value-name:"[HOST]:PORT" default:":9550"`
	MetricsPath   string `short:"m" long:"metrics-path" description:"Metrics path" value-name:"PATH" default:"/scrape"`
	V2RayEndpoint string `short:"e" long:"v2ray-endpoint" description:"V2Ray API endpoint" value-name:"HOST:PORT" default:"127.0.0.1:8080"`
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

	fmt.Printf("V2Ray Exporter %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)

	if opts.Version {
		os.Exit(0)
	}

	exporter = NewExporter(opts.V2RayEndpoint)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/scrape", scrapeHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
<head><title>V2Ray Exporter</title></head>
<body>
<h1>V2Ray Exporter ` + buildVersion + `</h1>
<p><a href='/metrics'>Exporter Metrics</a></p>
<p><a href='` + opts.MetricsPath + `'>Scrape V2Ray Metrics</a></p>
</body>
</html>
`))
		if err != nil {
			log.Debugf("Write() err: %s", err)
		}
	})

	log.Infof("Server is ready to handle incoming scrape requests.")
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
