package worker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/chunker"
	"ai-code-reviewer/internal/dedup"
	"ai-code-reviewer/internal/diff"
	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/observability"
	"ai-code-reviewer/internal/retry"
	"ai-code-reviewer/internal/review"
)

type Processor struct {
	queue    Queue
	client   github.Client
	comments github.CommentClient
	dedup    dedup.Store
	logger   *observability.Logger
	chunker  *chunker.Chunker
	ai       ai.Provider
}

func NewProcessor(
	q Queue,
	c github.Client,
	comments github.CommentClient,
	d dedup.Store,
	l *observability.Logger,
	a ai.Provider,
) *Processor {

	return &Processor{
		queue:    q,
		client:   c,
		comments: comments,
		dedup:    d,
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
		p.logger.Error("get files failed", "err", err)
		return
	}

	for _, f := range files {

		parsed, _ := diff.Parse(f.Patch)

		for _, pf := range parsed {

			content := pf.ToAIContext()

			chunks := p.chunker.Split(pf.Filename, content)

			for _, ch := range chunks {

				reviewText, err :=
					p.ai.Review(ctx, ai.ReviewRequest{
						File:    ch.File,
						Content: ch.Content,
					})

				if err != nil {
					p.logger.Error("ai failed", "err", err)
					continue
				}

				issues := review.ExtractIssues(reviewText)

				for _, is := range issues {

					// Create unique key
					key := fmt.Sprintf(
						"%s:%d:%s",
						ch.File,
						is.Line,
						hash(is.Text),
					)

					// Dedup check
					if p.dedup.Seen(ctx, key) {
						continue
					}

					comment := github.LineComment{
						Body: is.Text,
						Path: ch.File,
						Line: is.Line,
						Side: "RIGHT",
					}

					err = retry.Do(ctx, 3, time.Second, func() error {
						return p.comments.CreateLineComment(
							ctx, j.Repo, j.PR, comment,
						)
					})

					if err != nil {
						p.logger.Error("comment failed",
							"err", err,
						)
						continue
					}

					// mark as posted
					p.dedup.Mark(ctx, key)
				}

				p.logger.Info("AI REVIEW",
					"file", ch.File,
					"review", reviewText,
				)
			}
		}
	}
}

func hash(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}
