package review

import (
	"fmt"
	"strings"
)

type Issue struct {
	Line int
	Text string
}

func ExtractIssues(aiText string) []Issue {

	var issues []Issue

	for _, l := range strings.Split(aiText, "\n") {

		// Expect format: LINE: 23: message
		if strings.HasPrefix(l, "LINE:") {

			parts := strings.SplitN(l, ":", 3)
			if len(parts) < 3 {
				continue
			}

			// naive parse
			line := 0
			fmt.Sscanf(parts[1], "%d", &line)

			issues = append(issues, Issue{
				Line: line,
				Text: parts[2],
			})
		}
	}

	return issues
}
