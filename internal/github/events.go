package github

import "time"

type PullRequestEvent struct {
	Action      string      `json:"action"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
}

type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Head   struct {
		SHA string `json:"sha"`
	} `json:"head"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository struct {
	FullName string `json:"full_name"`
}
