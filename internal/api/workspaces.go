package api

import "net/url"

// Workspace is a subset of the workspace resource nested under each
// item's "workspace" key in GET /user/workspaces. Note: unlike the old
// (removed) /workspaces endpoint, this one does not return a "name" —
// only Slug and UUID are populated in practice.
type Workspace struct {
	UUID string `json:"uuid"`
	Slug string `json:"slug"`
	Name string `json:"name,omitempty"`
}

// workspaceAccess is the shape of each item in GET /user/workspaces's
// "values" array: the workspace is nested under a "workspace" key
// alongside the caller's access level (e.g. "administrator").
type workspaceAccess struct {
	Workspace Workspace `json:"workspace"`
}

// ListWorkspaces returns the workspaces the authenticated user belongs
// to. limit <= 0 fetches every page.
//
// This uses GET /user/workspaces, not the old GET /workspaces (removed by
// Atlassian's CHANGE-2770: cross-workspace listing without a user scope
// was sunset in favor of this endpoint). Requires the
// read:workspace:bitbucket API token scope.
func (c *Client) ListWorkspaces(limit int) ([]Workspace, error) {
	items, err := getPaginated[workspaceAccess](c, "/user/workspaces", limit)
	if err != nil {
		return nil, err
	}
	workspaces := make([]Workspace, len(items))
	for i, item := range items {
		workspaces[i] = item.Workspace
	}
	return workspaces, nil
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
