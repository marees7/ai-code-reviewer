package app

import (
	"net/http"

	"ai-code-reviewer/internal/github"
)

func (s *Server) routes() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.health)

	// GitHub webhook
	ghClient := github.NewClient(s.cfg, s.logger)

	gh := github.NewWebhookHandler(s.cfg, s.logger, ghClient)
	mux.HandleFunc("/webhook/github", gh.Handle)

	s.http.Handler = mux
}
