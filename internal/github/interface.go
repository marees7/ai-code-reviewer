package github

import "context"

type Client interface {
	GetPRFiles(ctx context.Context, repo string, pr int) ([]PRFile, error)
	GetPRDiff(ctx context.Context, repo string, pr int) (string, error)
	CreateComment(ctx context.Context, repo string, pr int, body string) error
}
