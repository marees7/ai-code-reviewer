package worker

import (
	"context"
	"strings"
	"testing"

	"ai-code-reviewer/internal/ai"
	"ai-code-reviewer/internal/budget"
	"ai-code-reviewer/internal/config"
	"ai-code-reviewer/internal/dedup"
	"ai-code-reviewer/internal/github"
	"ai-code-reviewer/internal/mocks"
	"ai-code-reviewer/internal/observability"
	"ai-code-reviewer/internal/ratelimit"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type clientStub struct {
	files []github.PRFile
}

func (c *clientStub) GetPRFiles(ctx context.Context, repo string, pr int) ([]github.PRFile, error) {
	return c.files, nil
}

func (c *clientStub) GetPRDiff(ctx context.Context, repo string, pr int) (string, error) {
	return "", nil
}

func (c *clientStub) CreateComment(ctx context.Context, repo string, pr int, body string) error {
	return nil
}

func (c *clientStub) CreateLineComment(ctx context.Context, repo string, pr int, comment github.LineComment) error {
	return nil
}

func TestFormatSummaryComment_NoIssues(t *testing.T) {
	body := formatSummaryComment(reviewSummary{
		TotalIssues:      0,
		PostedComments:   0,
		SeverityCounters: map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0},
	})

	require.Contains(t, body, "No issues detected")
}

func TestProcessorHandle_PostsSummaryComment(t *testing.T) {
	provider := mocks.NewProvider(t)
	comments := mocks.NewCommentClient(t)
	client := &clientStub{
		files: []github.PRFile{
			{
				Filename: "main.go",
				Patch: "diff --git a/main.go b/main.go\n" +
					"--- a/main.go\n" +
					"+++ b/main.go\n" +
					"@@ -1,1 +1,2 @@\n" +
					"-old\n" +
					"+new\n",
			},
		},
	}

	provider.
		EXPECT().
		Review(mock.Anything, mock.Anything).
		Return(
			ai.ReviewResponse{
				Content:  `{"issues":[{"line":1,"severity":"high","title":"nil check","suggestion":"add nil check"},{"line":2,"severity":"low","title":"style","suggestion":"rename var"}]}`,
				Provider: "openai",
				Model:    "gpt-3.5-turbo",
				Usage: ai.Usage{
					PromptTokens:     100,
					CompletionTokens: 80,
					TotalTokens:      180,
				},
			},
			nil,
		).
		Once()

	comments.
		EXPECT().
		CreateLineComment(mock.Anything, "acme/repo", 7, mock.Anything).
		Return(nil).
		Twice()

	comments.
		EXPECT().
		CreateComment(mock.Anything, "acme/repo", 7, mock.MatchedBy(func(body string) bool {
			return strings.Contains(body, "Total issues found: 2") &&
				strings.Contains(body, "Line comments posted: 2") &&
				strings.Contains(body, "Estimated cost (USD):") &&
				strings.Contains(body, "High: 1") &&
				strings.Contains(body, "Low: 1")
		})).
		Return(nil).
		Once()

	p := NewProcessor(
		NewMemoryQueue(1),
		client,
		comments,
		dedup.NewMemory(),
		observability.NewLogger(&config.Config{LogLevel: "info"}),
		provider,
		ratelimit.New(100, 100),
		nil,
	)

	p.handle(context.Background(), Job{Repo: "acme/repo", PR: 7})
}

func TestProcessorHandle_StopsWhenBudgetExceeded(t *testing.T) {
	provider := mocks.NewProvider(t)
	comments := mocks.NewCommentClient(t)
	client := &clientStub{
		files: []github.PRFile{
			{
				Filename: "a.go",
				Patch: "diff --git a/a.go b/a.go\n" +
					"--- a/a.go\n" +
					"+++ b/a.go\n" +
					"@@ -1,1 +1,2 @@\n" +
					"-old\n" +
					"+new\n",
			},
			{
				Filename: "b.go",
				Patch: "diff --git a/b.go b/b.go\n" +
					"--- a/b.go\n" +
					"+++ b/b.go\n" +
					"@@ -1,1 +1,2 @@\n" +
					"-old\n" +
					"+new\n",
			},
		},
	}

	provider.
		EXPECT().
		Review(mock.Anything, mock.Anything).
		Return(
			ai.ReviewResponse{
				Content:  `{"issues":[]}`,
				Provider: "openai",
				Model:    "gpt-4o",
				Usage: ai.Usage{
					PromptTokens:     1000,
					CompletionTokens: 1000,
					TotalTokens:      2000,
				},
			},
			nil,
		).
		Once()

	comments.
		EXPECT().
		CreateComment(mock.Anything, "acme/repo", 9, mock.MatchedBy(func(body string) bool {
			return strings.Contains(body, "Budget guard triggered")
		})).
		Return(nil).
		Once()

	guard := budget.NewGuard(true, 100.0, 0.01, budget.NewMemoryStore())

	p := NewProcessor(
		NewMemoryQueue(1),
		client,
		comments,
		dedup.NewMemory(),
		observability.NewLogger(&config.Config{LogLevel: "info"}),
		provider,
		ratelimit.New(100, 100),
		guard,
	)

	p.handle(context.Background(), Job{Repo: "acme/repo", PR: 9})
}
