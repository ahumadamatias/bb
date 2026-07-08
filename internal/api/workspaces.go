package api

import "net/url"

// Workspace is a subset of the workspace resource returned by
// GET /workspaces.
type Workspace struct {
	UUID string `json:"uuid"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// ListWorkspaces returns the workspaces the authenticated user belongs
// to. limit <= 0 fetches every page.
func (c *Client) ListWorkspaces(limit int) ([]Workspace, error) {
	return getPaginated[Workspace](c, "/workspaces", limit)
}

// workspaceMembership is the shape of each item in
// GET /workspaces/{workspace}/members's "values" array.
type workspaceMembership struct {
	User Account `json:"user"`
}

// ListWorkspaceMembers returns the users belonging to a workspace, used
// to resolve `--reviewer` names to the UUIDs the API requires.
func (c *Client) ListWorkspaceMembers(workspace string) ([]Account, error) {
	memberships, err := getPaginated[workspaceMembership](c, "/workspaces/"+url.PathEscape(workspace)+"/members", 0)
	if err != nil {
		return nil, err
	}
	members := make([]Account, len(memberships))
	for i, m := range memberships {
		members[i] = m.User
	}
	return members, nil
}
