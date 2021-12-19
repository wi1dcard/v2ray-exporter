package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v4/app/stats/command"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Exporter struct {
	sync.Mutex
	endpoint           string
	scrapeTimeout      time.Duration
	registry           *prometheus.Registry
	totalScrapes       prometheus.Counter
	metricDescriptions map[string]*prometheus.Desc
}

func NewExporter(endpoint string, scrapeTimeout time.Duration) *Exporter {
	e := Exporter{
		endpoint:      endpoint,
		scrapeTimeout: scrapeTimeout,
		registry:      prometheus.NewRegistry(),

		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "v2ray",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes performed",
		}),
	}

	e.metricDescriptions = map[string]*prometheus.Desc{}

	for k, desc := range map[string]struct {
		txt  string
		lbls []string
	}{
		"up":                           {txt: "Indicate scrape succeeded or not"},
		"scrape_duration_seconds":      {txt: "Scrape duration in seconds"},
		"uptime_seconds":               {txt: "V2Ray uptime in seconds"},
		"traffic_uplink_bytes_total":   {txt: "Number of transmitted bytes", lbls: []string{"dimension", "target"}},
		"traffic_downlink_bytes_total": {txt: "Number of receieved bytes", lbls: []string{"dimension", "target"}},
	} {
		e.metricDescriptions[k] = e.newMetricDescr(k, desc.txt, desc.lbls)
	}

	e.registry.MustRegister(&e)

	return &e
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Lock()
	defer e.Unlock()
	e.totalScrapes.Inc()

	start := time.Now().UnixNano()

	var up float64 = 1
	if err := e.scrapeV2Ray(ch); err != nil {
		up = 0
		logrus.Warnf("Scrape failed: %s", err)
	}

	e.registerConstMetricGauge(ch, "up", up)
	e.registerConstMetricGauge(ch, "scrape_duration_seconds", float64(time.Now().UnixNano()-start)/1000000000)

	ch <- e.totalScrapes
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range e.metricDescriptions {
		ch <- desc
	}

	ch <- e.totalScrapes.Desc()
}

func (e *Exporter) scrapeV2Ray(ch chan<- prometheus.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.scrapeTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, e.endpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to dial: %w, timeout: %v", err, e.scrapeTimeout)
	}
	defer conn.Close()

	client := command.NewStatsServiceClient(conn)

	if err := e.scrapeV2RaySysMetrics(ctx, ch, client); err != nil {
		return err
	}

	if err := e.scrapeV2RayMetrics(ctx, ch, client); err != nil {
		return err
	}

	return nil
}

func (e *Exporter) scrapeV2RayMetrics(ctx context.Context, ch chan<- prometheus.Metric, client command.StatsServiceClient) error {
	resp, err := client.QueryStats(ctx, &command.QueryStatsRequest{Reset_: false})
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	for _, s := range resp.GetStat() {
		// example value: inbound>>>socks-proxy>>>traffic>>>uplink
		p := strings.Split(s.GetName(), ">>>")
		metric := p[2] + "_" + p[3] + "_bytes_total"
		dimension := p[0]
		target := p[1]

		e.registerConstMetricCounter(ch, metric, float64(s.GetValue()), dimension, target)
	}

	return nil
}

func (e *Exporter) scrapeV2RaySysMetrics(ctx context.Context, ch chan<- prometheus.Metric, client command.StatsServiceClient) error {
	resp, err := client.GetSysStats(ctx, &command.SysStatsRequest{})
	if err != nil {
		return fmt.Errorf("failed to get sys stats: %w", err)
	}

	e.registerConstMetricGauge(ch, "uptime_seconds", float64(resp.GetUptime()))

	// We followed the naming style of Go collector from Prometheus.
	// See: https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go
	e.registerConstMetricGauge(ch, "goroutines", float64(resp.GetNumGoroutine()))
	e.registerConstMetricGauge(ch, "memstats_alloc_bytes", float64(resp.GetAlloc()))
	e.registerConstMetricGauge(ch, "memstats_alloc_bytes_total", float64(resp.GetTotalAlloc()))
	e.registerConstMetricGauge(ch, "memstats_sys_bytes", float64(resp.GetSys()))
	e.registerConstMetricGauge(ch, "memstats_mallocs_total", float64(resp.GetMallocs()))
	e.registerConstMetricGauge(ch, "memstats_frees_total", float64(resp.GetFrees()))

	// The metric live_objects was removed. You may calculate it in Prometheus using:
	// memstats_live_objects_total = memstats_mallocs_total - memstats_frees_total
	// See: https://prometheus.io/docs/instrumenting/writing_exporters/#drop-less-useful-statistics

	// These metrics below are not directly exposed by Go collector.
	// Therefore we only add the "memstats_" prefix without changing their original names.
	e.registerConstMetricGauge(ch, "memstats_num_gc", float64(resp.GetNumGC()))
	e.registerConstMetricGauge(ch, "memstats_pause_total_ns", float64(resp.GetPauseTotalNs()))

	return nil
}

func (e *Exporter) registerConstMetricGauge(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	e.registerConstMetric(ch, metric, val, prometheus.GaugeValue, labels...)
}

func (e *Exporter) registerConstMetricCounter(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	e.registerConstMetric(ch, metric, val, prometheus.CounterValue, labels...)
}

func (e *Exporter) registerConstMetric(ch chan<- prometheus.Metric, metric string, val float64, valType prometheus.ValueType, labelValues ...string) {
	descr := e.metricDescriptions[metric]
	if descr == nil {
		descr = e.newMetricDescr(metric, metric+" metric", nil)
	}

	if m, err := prometheus.NewConstMetric(descr, valType, val, labelValues...); err == nil {
		ch <- m
	} else {
		logrus.Debugf("NewConstMetric() err: %s", err)
	}
}

func (e *Exporter) newMetricDescr(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName("v2ray", "", metricName), docString, labels, nil)
}
