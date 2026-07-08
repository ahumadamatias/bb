package api

import (
	"net/http/httptest"
	"testing"
)

func TestListWorkspaces(t *testing.T) {
	server := httptest.NewServer(serveFixture(t, "testdata/workspaces.json"))
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
}
