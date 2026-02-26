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

	AITokens = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_reviewer_ai_tokens_total",
			Help: "Total AI tokens",
		},
		[]string{"provider", "model", "type"},
	)

	AICostUSD = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_reviewer_ai_cost_usd_total",
			Help: "Total estimated AI cost in USD",
		},
		[]string{"provider", "model"},
	)

	AIBudgetBlocks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_reviewer_budget_block_total",
			Help: "Total budget block events",
		},
		[]string{"scope"},
	)
)

func InitMetrics() {
	registerMetricsOnce.Do(func() {
		prometheus.MustRegister(AICalls, AIErrors, AILatency, AITokens, AICostUSD, AIBudgetBlocks)
	})
}
