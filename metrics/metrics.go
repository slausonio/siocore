package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)
	HttpRequestFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_failures_total",
			Help: "Total number of failed HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)
)

func InitPrometheus() {
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestFailures)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		panic(err)
	}
}
