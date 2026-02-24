package github

type LineComment struct {
	Body string `json:"body"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"` // RIGHT = new code
}
