package diff

import (
	"strings"
)

func (f FileDiff) ToAIContext() string {

	var b strings.Builder

	b.WriteString("File: " + f.Filename + "\n\n")

	for _, h := range f.Hunks {

		b.WriteString("Hunk:\n")

		for _, l := range h.Lines {

			prefix := " "
			if l.Type == Added {
				prefix = "+"
			}
			if l.Type == Removed {
				prefix = "-"
			}

			b.WriteString(
				prefix + l.Content + "\n",
			)
		}
	}

	return b.String()
}
