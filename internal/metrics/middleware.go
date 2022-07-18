package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{w, http.StatusOK}
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.statusCode = statusCode
	mrw.ResponseWriter.WriteHeader(statusCode)
}

func RequestsMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mrw := newMetricsResponseWriter(w)
		t1 := time.Now()
		next.ServeHTTP(mrw, r)
		t2 := time.Now()
		duration := t2.Sub(t1)

		metrics.RequestsTotal.With(prometheus.Labels{
			"method":  r.Method,
			"handler": r.RequestURI,
			"status":  strconv.Itoa(mrw.statusCode),
		}).Inc()
		metrics.RequestsDuration.With(prometheus.Labels{
			"method":  r.Method,
			"handler": r.RequestURI,
			"status":  strconv.Itoa(mrw.statusCode),
		}).Observe(float64(duration.Milliseconds()))
	})
}
