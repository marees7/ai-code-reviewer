package diff

type FileDiff struct {
	Filename string
	Hunks    []Hunk
}

type Hunk struct {
	OldStart int
	NewStart int
	Lines    []Line
}

type Line struct {
	Type      LineType
	Content   string
	OldNumber int
	NewNumber int
}

type LineType string

const (
	Added   LineType = "added"
	Removed LineType = "removed"
	Context LineType = "context"
)
