package worker

import (
	"context"
	"time"

	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/observability"
)

type Processor struct {
	queue  Queue
	client github.Client
	logger *observability.Logger
}

func NewProcessor(q Queue, c github.Client, l *observability.Logger) *Processor {
	return &Processor{
		queue:  q,
		client: c,
		logger: l,
	}
}

func (p *Processor) Start(ctx context.Context) {

	go func() {
		for {
			job, err := p.queue.Pop(ctx)
			if err != nil {
				continue
			}

			p.handle(job)
		}
	}()
}

func (p *Processor) handle(j Job) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	files, err := p.client.GetPRFiles(ctx, j.Repo, j.PR)
	if err != nil {
		p.logger.Error("worker failed", "error", err)
		return
	}

	p.logger.Info("worker processing",
		"repo", j.Repo,
		"pr", j.PR,
		"files", len(files),
	)
}
