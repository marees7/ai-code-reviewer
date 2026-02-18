package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/config"
	"ai-code-reviewer/internal/observability"
)

type Server struct {
	cfg    *config.Config
	logger *observability.Logger
	openAI *ai.OpenAI
	http   *http.Server
}

func NewServer(cfg *config.Config, logger *observability.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
		openAI: ai.NewOpenAI(
			cfg.OpenAIKey,
			cfg.OpenAIModel,
		),
	}

	s.http = &http.Server{
		Addr:         ":" + cfg.Port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
	}
	s.routes()

	return s
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.http.Shutdown(context.Background())
	}()

	s.logger.Info("starting server",
		"port", s.cfg.Port,
		"env", s.cfg.Env,
	)

	if err := s.http.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}

	return nil
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
