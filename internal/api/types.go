package api

import "net/url"

// Account represents a Bitbucket user or team as embedded in other
// resources (PR author, reviewer, comment author, etc).
type Account struct {
	UUID        string `json:"uuid,omitempty"`
	Username    string `json:"username,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// Links is the common "links" envelope Bitbucket attaches to most
// resources. Only the fields bb actually uses are modeled.
type Links struct {
	HTML struct {
		Href string `json:"href"`
	} `json:"html"`
}

func repoPath(workspace, repoSlug string) string {
	return "/repositories/" + url.PathEscape(workspace) + "/" + url.PathEscape(repoSlug)
}
