package diff

import (
	"regexp"
	"strconv"
)

var hunkRe = regexp.MustCompile(`@@ -(\d+),?\d* \+(\d+),?\d* @@`)

func parseHunkHeader(line string) Hunk {

	m := hunkRe.FindStringSubmatch(line)

	oldStart, _ := strconv.Atoi(m[1])
	newStart, _ := strconv.Atoi(m[2])

	return Hunk{
		OldStart: oldStart,
		NewStart: newStart,
	}
}
