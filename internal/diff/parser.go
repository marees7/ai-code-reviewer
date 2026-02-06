package diff

import (
	"bufio"
	"strings"
)

func Parse(patch string) ([]FileDiff, error) {

	var files []FileDiff
	var current *FileDiff
	var hunk *Hunk

	scanner := bufio.NewScanner(strings.NewReader(patch))

	for scanner.Scan() {
		line := scanner.Text()

		// New file
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				files = append(files, *current)
			}

			current = &FileDiff{}
			continue
		}

		// Filename
		if strings.HasPrefix(line, "+++ b/") {
			if current != nil {
				current.Filename = strings.TrimPrefix(line, "+++ b/")
			}
			continue
		}

		// Hunk start
		if strings.HasPrefix(line, "@@") {
			if current == nil {
				continue
			}

			h := parseHunkHeader(line)
			hunk = &h
			current.Hunks = append(current.Hunks, *hunk)
			continue
		}

		// Content lines
		if hunk != nil {
			l := parseLine(line, hunk)
			hunk.Lines = append(hunk.Lines, l)
		}
	}

	if current != nil {
		files = append(files, *current)
	}

	return files, nil
}
