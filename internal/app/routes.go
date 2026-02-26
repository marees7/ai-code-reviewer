package app

import (
	"context"
	"net/http"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/budget"
	"ai-code-reviewer/internal/dedup"
	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/observability"
	"ai-code-reviewer/internal/ratelimit"
	"ai-code-reviewer/internal/worker"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	healthPath        = "/health"
	metricsPath       = "/metrics"
	githubWebhookPath = "/webhook/github"
)

func (s *Server) routes() {
	if s.http == nil {
		return
	}

	mux := http.NewServeMux()

	mux.HandleFunc(healthPath, s.health)
	// create core queue

	// create queue based on config
	queue := worker.NewQueue(s.cfg)

	// adapter so github pkg doesn't know worker
	adapter := worker.NewAdapter(queue)

	// github client
	ghClient := github.NewClient(s.cfg, s.logger)

	// webhook
	gh := github.NewWebhookHandler(
		s.cfg,
		s.logger,
		adapter,
	)

	provider := ai.NewProvider(s.cfg)

	// add circuit breaker
	providerWithCB := ai.NewCircuitBreaker(provider)

	// optional fallback still works
	fallback := ai.NewFallback(
		providerWithCB,
		ai.NewOllama(
			s.cfg.OllamaURL,
			s.cfg.OllamaModel,
		),
	)

	dedup := dedup.NewMemory()

	//ratelimiter
	rateLimiter := ratelimit.New(s.cfg.RateLimitRPS, s.cfg.RateLimitBurst)

	// background processor
	processor := worker.NewProcessor(
		queue,
		ghClient,
		ghClient,
		dedup,
		s.logger,
		fallback,
		rateLimiter,
		budget.NewGuard(
			s.cfg.BudgetEnabled,
			s.cfg.BudgetDailyUSD,
			s.cfg.BudgetPerPRUSD,
			budget.NewMemoryStore(),
		),
	)

	// init metrics
	observability.InitMetrics()

	mux.HandleFunc(githubWebhookPath, gh.Handle)
	mux.Handle(metricsPath, promhttp.Handler())

	processor.Start(context.Background())

	s.http.Handler = mux
}
