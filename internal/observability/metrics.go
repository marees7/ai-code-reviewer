package observability

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerMetricsOnce sync.Once

	AICalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_reviewer_ai_calls_total",
			Help: "Total AI calls",
		},
		[]string{"provider"},
	)

	AIErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_reviewer_ai_errors_total",
			Help: "Total AI errors",
		},
		[]string{"provider"},
	)

	AILatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_reviewer_ai_latency_seconds",
			Help:    "AI call latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider"},
	)
)

func InitMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(AICalls, AIErrors, AILatency)
	})
}
