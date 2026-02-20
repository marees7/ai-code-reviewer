package worker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/chunker"
	"ai-code-reviewer/internal/dedup"
	"ai-code-reviewer/internal/diff"
	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/observability"
	"ai-code-reviewer/internal/ratelimit"
	"ai-code-reviewer/internal/retry"
	"ai-code-reviewer/internal/review"
)

type Processor struct {
	queue       Queue
	client      github.Client
	comments    github.CommentClient
	dedup       dedup.Store
	logger      *observability.Logger
	chunker     *chunker.Chunker
	ai          ai.Provider
	rateLimiter *ratelimit.Limiter
}

func NewProcessor(
	q Queue,
	c github.Client,
	comments github.CommentClient,
	d dedup.Store,
	l *observability.Logger,
	a ai.Provider,
	rl *ratelimit.Limiter,
) *Processor {

	return &Processor{
		queue:       q,
		client:      c,
		comments:    comments,
		dedup:       d,
		logger:      l,
		chunker:     chunker.New(3000),
		ai:          a,
		rateLimiter: rl,
	}
}

func (p *Processor) Start(ctx context.Context) {

	go func() {
		for {
			job, err := p.queue.Pop(ctx)
			if err != nil {
				if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
					return
				}
				continue
			}

			p.handle(ctx, job)
		}
	}()
}

func (p *Processor) handle(parent context.Context, j Job) {

	ctx, cancel := context.WithTimeout(
		parent,
		90*time.Second,
	)
	defer cancel()

	files, err := p.client.GetPRFiles(ctx, j.Repo, j.PR)
	if err != nil {
		p.logger.Error("get files failed", "err", err)
		return
	}

	limiter := p.rateLimiter.Get(j.Repo)

	for _, f := range files {

		parsed, err := diff.Parse(f.Patch)
		if err != nil {
			p.logger.Error("diff parse failed", "file", f.Filename, "err", err)
			continue
		}

		for _, pf := range parsed {

			content := pf.ToAIContext()

			chunks := p.chunker.Split(pf.Filename, content)

			for _, ch := range chunks {

				err := limiter.Wait(ctx)
				if err != nil {
					p.logger.Error("rate limiter error", "err", err)
					return
				}

				startTime := time.Now()

				reviewText, err :=
					p.ai.Review(ctx, ai.ReviewRequest{
						File:    ch.File,
						Content: ch.Content,
					})

				duration := time.Since(startTime).Seconds()

				//Later we can make this as dynamic
				provider := "primary"

				observability.AICalls.WithLabelValues(provider).Inc()
				observability.AILatency.WithLabelValues(provider).Observe(duration)

				if err != nil {
					observability.AIErrors.WithLabelValues(provider).Inc()
					p.logger.Error("ai failed", "err", err)
					continue
				}

				result, err := review.ParseResult(reviewText)
				if err != nil {
					p.logger.Error("invalid ai json",
						"err", err,
					)
					continue
				}

				for _, is := range result.Issues {

					// Create unique key
					key := fmt.Sprintf(
						"%s:%d:%s",
						ch.File,
						is.Line,
						hash(is.Severity+is.Title+is.Suggestion),
					)

					// Dedup check
					if p.dedup.Seen(ctx, key) {
						continue
					}

					comment := github.LineComment{
						Body: is.Suggestion,
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
