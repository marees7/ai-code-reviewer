package worker_test

import (
	"context"
	"testing"

	"ai-code-reviewer/internal/worker"

	"github.com/stretchr/testify/suite"
)

type RedisSuite struct {
	suite.Suite
	q *worker.RedisQueue
}

func (s *RedisSuite) SetupSuite() {
	s.q = worker.NewRedisQueue("localhost:6379", "test")
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
