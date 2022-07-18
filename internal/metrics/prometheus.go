package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestsDuration *prometheus.HistogramVec
}

var metrics = Metrics{
	RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "go_app_requests_total",
		Help: "The total number of processed requests",
	}, []string{"method", "handler", "status"}),
	RequestsDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "go_app_requests_duration",
	}, []string{"method", "handler", "status"}),
}

func InitMetrics(reg *prometheus.Registry) {
	reg.MustRegister(metrics.RequestsTotal)
	reg.MustRegister(metrics.RequestsDuration)
}
