package worker

import "context"

// This is the INTERNAL queue abstraction
type Queue interface {
	Push(ctx context.Context, j Job) error
	Pop(ctx context.Context) (Job, error)
}

type Job struct {
	Repo string
	PR   int
}
