package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_events_total",
			Help: "Total number of events received",
		},
		[]string{"device", "type", "severity"},
	)

	criticalErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "collector_critical_errors_total",
			Help: "Total number of critical errors",
		},
		[]string{"device"},
	)
)

func initMetrics() {
	prometheus.MustRegister(totalEvents, criticalErrors)
}

func recordMetrics(e Event) {
	totalEvents.WithLabelValues(e.Device, e.Type, e.Severity).Inc()
	if e.Severity == "critical" {
		criticalErrors.WithLabelValues(e.Device).Inc()
	}
}

func promHTTPHandler() http.Handler {
	return promhttp.Handler()
}
