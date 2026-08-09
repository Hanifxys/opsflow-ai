package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"service", "method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal, HTTPRequestDuration)
}

// MetricsMiddleware returns an HTTP middleware that records Prometheus metrics.
func MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			statusStr := strconv.Itoa(rw.statusCode)
			path := r.URL.Path

			HTTPRequestsTotal.WithLabelValues(serviceName, r.Method, path, statusStr).Inc()
			HTTPRequestDuration.WithLabelValues(serviceName, r.Method, path).Observe(duration)
		})
	}
}

// Handler returns the Prometheus HTTP metrics handler for /metrics endpoints.
func Handler() http.Handler {
	return promhttp.Handler()
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
