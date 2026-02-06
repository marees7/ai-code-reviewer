package github

type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

type PullRequest struct {
	Number int `json:"number"`

	// 🔥 THIS WAS MISSING
	Draft bool `json:"draft"`

	User struct {
		Login string `json:"login"`
	} `json:"user"`

	Head struct {
		SHA string `json:"sha"`
	} `json:"head"`

	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`

	Title string `json:"title"`
	Body  string `json:"body"`
}

type Repository struct {
	FullName string `json:"full_name"`
}
