// Package metrics defines the platform's business metrics on the default
// Prometheus registry. Every service exposes them via health.Serve's /metrics.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ChecksTotal counts completed health checks by resulting status.
	ChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sitemon_checks_total",
		Help: "Health checks run, by resulting status.",
	}, []string{"status"})

	// CheckDuration observes how long a health check took.
	CheckDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "sitemon_check_duration_seconds",
		Help:    "Health check duration in seconds.",
		Buckets: prometheus.DefBuckets,
	})

	// JobsDispatched counts check jobs the scheduler published.
	JobsDispatched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sitemon_jobs_dispatched_total",
		Help: "Check jobs dispatched by the scheduler.",
	})

	// AlertsTotal counts alerts fired by type (down/recovery).
	AlertsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sitemon_alerts_total",
		Help: "Alerts fired, by type.",
	}, []string{"type"})
)

// ObserveCheck records a single check's status and duration.
func ObserveCheck(status string, d time.Duration) {
	ChecksTotal.WithLabelValues(status).Inc()
	CheckDuration.Observe(d.Seconds())
}
