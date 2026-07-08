package api

import (
	"net/http"
	"net/url"
	"time"
)

// Repository is a subset of the repository resource returned by the
// /repositories endpoints.
type Repository struct {
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	UpdatedOn   time.Time `json:"updated_on"`
	Mainbranch  struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
	Links Links `json:"links"`
}

// ListRepositories lists the repositories in a workspace, most recently
// updated first. limit <= 0 fetches every page.
func (c *Client) ListRepositories(workspace string, limit int) ([]Repository, error) {
	path := "/repositories/" + url.PathEscape(workspace) + "?sort=-updated_on"
	return getPaginated[Repository](c, path, limit)
}

// GetRepository fetches a single repository, used to discover its main
// branch (the default `--dest` for `bb pr create`).
func (c *Client) GetRepository(workspace, repoSlug string) (*Repository, error) {
	var repo Repository
	if err := c.doJSON(http.MethodGet, repoPath(workspace, repoSlug), nil, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}
