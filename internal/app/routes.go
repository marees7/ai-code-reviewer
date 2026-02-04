package app

import (
	"net/http"

	"ai-code-reviewer/internal/github"
)

func (s *Server) routes() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.health)

	// GitHub webhook
	gh := github.NewWebhookHandler(s.cfg, s.logger)
	mux.HandleFunc("/webhook/github", gh.Handle)

	s.http.Handler = mux
}
