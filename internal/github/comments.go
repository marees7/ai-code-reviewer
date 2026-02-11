package github

import "context"

//go:generate mockery --name CommentClient --output ../mocks --with-expecter
type CommentClient interface {
	CreateLineComment(ctx context.Context, repo string, pr int, comment LineComment) error
}
