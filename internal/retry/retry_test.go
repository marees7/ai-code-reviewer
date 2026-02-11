package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-code-reviewer/internal/retry"

	"github.com/stretchr/testify/suite"
)

type RetrySuite struct {
	suite.Suite
}

func (s *RetrySuite) Test_Eventual_Success() {

	calls := 0

	err := retry.Do(
		context.Background(),
		3,
		1*time.Millisecond,
		func() error {
			calls++
			if calls < 2 {
				return errors.New("fail")
			}
			return nil
		},
	)

	s.NoError(err)
	s.Equal(2, calls)
}

func TestRetrySuite(t *testing.T) {
	suite.Run(t, new(RetrySuite))
}
