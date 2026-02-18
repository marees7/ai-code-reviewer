package worker_test

import (
	"testing"

	"ai-code-reviewer/internal/dedup"
	"ai-code-reviewer/internal/mocks"
	"ai-code-reviewer/internal/ratelimit"
	"ai-code-reviewer/internal/worker"

	"github.com/stretchr/testify/suite"
)

type ProcessorSuite struct {
	suite.Suite

	ai        *mocks.Provider
	comments  *mocks.CommentClient
	queue     *worker.MemoryQueue
	processor *worker.Processor
}

func (s *ProcessorSuite) SetupTest() {

	s.ai = mocks.NewProvider(s.T())
	s.comments = mocks.NewCommentClient(s.T())
	s.queue = worker.NewMemoryQueue(10)

	s.processor = worker.NewProcessor(
		s.queue,
		nil,
		s.comments,
		dedup.NewMemory(),
		nil,
		s.ai,
		ratelimit.New(1, 1),
	)
}

func TestProcessorSuite(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}
