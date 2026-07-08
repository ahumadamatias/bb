package api

import "time"

// Branch is a subset of the branch resource returned by
// /repositories/{ws}/{repo}/refs/branches.
type Branch struct {
	Name   string `json:"name"`
	Target struct {
		Hash   string    `json:"hash"`
		Date   time.Time `json:"date"`
		Author struct {
			Raw  string   `json:"raw"`
			User *Account `json:"user"`
		} `json:"author"`
	} `json:"target"`
}

// ListBranches lists the branches of a repository. limit <= 0 fetches
// every page.
func (c *Client) ListBranches(workspace, repoSlug string, limit int) ([]Branch, error) {
	path := repoPath(workspace, repoSlug) + "/refs/branches"
	return getPaginated[Branch](c, path, limit)
}
