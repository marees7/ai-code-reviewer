package github

import "context"

type CommentClient interface {
	CreateComment(ctx context.Context, repo string, pr int, body string) error
}
