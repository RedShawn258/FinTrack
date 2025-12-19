package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus metrics for outbox events
type Metrics struct {
	PendingTotal   prometheus.Gauge
	SentTotal      prometheus.Counter
	FailedTotal    prometheus.Counter
	RetriesTotal   prometheus.Counter
	PublishLatency prometheus.Histogram
}

// NewMetrics creates and registers Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		PendingTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "outbox_pending_total",
			Help: "Number of pending outbox events",
		}),
		SentTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "outbox_sent_total",
			Help: "Total number of outbox events successfully published",
		}),
		FailedTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "outbox_failed_total",
			Help: "Total number of outbox events that failed after max retries",
		}),
		RetriesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "outbox_retries_total",
			Help: "Total number of outbox event retry attempts",
		}),
		PublishLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "outbox_publish_latency_seconds",
			Help:    "Latency of publishing outbox events in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
		}),
	}
}
