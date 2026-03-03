package github

import "context"

// Webhook only knows THIS interface
type JobQueue interface {
	Enqueue(ctx context.Context, tenant, repo string, pr int) error
}
