package worker_test

import (
	"context"
	"testing"
	"time"

	"ai-code-reviewer/internal/worker"

	"github.com/stretchr/testify/suite"
)

type RedisSuite struct {
	suite.Suite
	q *worker.RedisQueue
}

func (s *RedisSuite) SetupSuite() {
	s.q = worker.NewRedisQueue("localhost:6379", "test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.q.Push(ctx, worker.Job{Repo: "health/check", PR: 0}); err != nil {
		s.T().Skip("redis unavailable on localhost:6379")
	}
	_, _ = s.q.Pop(ctx)
}

func (s *RedisSuite) TestPushPop() {

	ctx := context.Background()

	job := worker.Job{Repo: "a/b", PR: 1}

	err := s.q.Push(ctx, job)
	s.NoError(err)

	out, err := s.q.Pop(ctx)

	s.NoError(err)
	s.Equal(job.Repo, out.Repo)
}

func TestRedis(t *testing.T) {
	suite.Run(t, new(RedisSuite))
}
