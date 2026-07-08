package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWorkspaces(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		serveFixture(t, "testdata/workspaces.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	workspaces, err := c.ListWorkspaces(0)
	if err != nil {
		t.Fatalf("ListWorkspaces() error = %v", err)
	}
	if len(workspaces) != 1 || workspaces[0].Slug != "myworkspace" {
		t.Errorf("workspaces = %+v", workspaces)
	}
	// GET /workspaces (cross-workspace listing without a specific user
	// scope) was removed by Atlassian's CHANGE-2770; regression guard
	// against reintroducing that path.
	if want := "/user/workspaces"; gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
}
