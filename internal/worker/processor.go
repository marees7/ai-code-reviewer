package worker

import (
	"context"
	"time"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/chunker"
	"ai-code-reviewer/internal/diff"
	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/observability"
	"ai-code-reviewer/internal/review"
)

type Processor struct {
	queue    Queue
	client   github.Client
	comments github.CommentClient
	logger   *observability.Logger
	chunker  *chunker.Chunker
	ai       ai.Provider
}

func NewProcessor(
	q Queue,
	c github.Client,
	comments github.CommentClient,
	l *observability.Logger,
	a ai.Provider,
) *Processor {

	return &Processor{
		queue:    q,
		client:   c,
		comments: comments,
		logger:   l,
		chunker:  chunker.New(3000),
		ai:       a,
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

	ctx, cancel := context.WithTimeout(
		context.Background(),
		90*time.Second,
	)
	defer cancel()

	files, err := p.client.GetPRFiles(ctx, j.Repo, j.PR)
	if err != nil {
		return
	}

	for _, f := range files {

		parsed, _ := diff.Parse(f.Patch)

		for _, pf := range parsed {

			content := pf.ToAIContext()

			chunks := p.chunker.Split(
				pf.Filename,
				content,
			)

			for _, ch := range chunks {

				reviewText, err := p.ai.Review(
					ctx,
					ai.ReviewRequest{
						File:    ch.File,
						Content: ch.Content,
					},
				)

				if err != nil {
					continue
				}

				// Format
				comment := review.FormatComment(
					ch.File,
					reviewText,
				)

				// Post to GitHub
				err = p.comments.CreateComment(
					ctx,
					j.Repo,
					j.PR,
					comment,
				)

				if err != nil {
					p.logger.Error("comment failed",
						"error", err,
					)
				}

				p.logger.Info("AI REVIEW",
					"file", ch.File,
					"review", reviewText,
				)
			}
		}
	}
}
