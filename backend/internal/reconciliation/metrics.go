package reconciliation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus metrics for reconciliation
type Metrics struct {
	ReconciliationFailures prometheus.Counter
	ReconciliationTotal    prometheus.Counter
	ReconciliationDuration prometheus.Histogram
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		ReconciliationFailures: promauto.NewCounter(prometheus.CounterOpts{
			Name: "reconciliation_failures_total",
			Help: "Total number of account balance reconciliation failures",
		}),
		ReconciliationTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "reconciliation_checks_total",
			Help: "Total number of reconciliation checks performed",
		}),
		ReconciliationDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "reconciliation_duration_seconds",
			Help:    "Duration of reconciliation checks in seconds",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 10), // 10ms to ~10s
		}),
	}
}
