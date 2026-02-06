package app

import (
	"context"
	"net/http"

	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/worker"
)

func (s *Server) routes() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.health)
	// create core queue
	memQueue := worker.NewMemoryQueue(100)

	// adapter so github pkg doesn't know worker
	adapter := worker.NewAdapter(memQueue)

	// github client
	ghClient := github.NewClient(s.cfg, s.logger)

	// webhook
	gh := github.NewWebhookHandler(
		s.cfg,
		s.logger,
		ghClient,
		adapter,
	)

	// background processor
	processor := worker.NewProcessor(
		memQueue,
		ghClient,
		s.logger,
	)

	mux.HandleFunc("/webhook/github", gh.Handle)

	processor.Start(context.Background())

	s.http.Handler = mux
}
