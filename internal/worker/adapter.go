package worker

import "context"

// Adapter implements github.JobQueue
type Adapter struct {
	q Queue
}

func NewAdapter(q Queue) *Adapter {
	return &Adapter{q: q}
}

func (a *Adapter) Enqueue(ctx context.Context, tenant, repo string, pr int) error {
	return a.q.Push(ctx, Job{
		Tenant: tenant,
		Repo:   repo,
		PR:     pr,
	})
}
