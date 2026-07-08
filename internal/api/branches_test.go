package api

import (
	"net/http/httptest"
	"testing"
)

func TestListBranches(t *testing.T) {
	server := httptest.NewServer(serveFixture(t, "testdata/branches.json"))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	branches, err := c.ListBranches("myworkspace", "bb", 0)
	if err != nil {
		t.Fatalf("ListBranches() error = %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("len(branches) = %d, want 2", len(branches))
	}
	if branches[0].Name != "main" || branches[0].Target.Hash != "abc1234" {
		t.Errorf("branches[0] = %+v", branches[0])
	}
	if branches[0].Target.Author.User == nil || branches[0].Target.Author.User.DisplayName != "Matias Ahumada" {
		t.Errorf("branches[0].Target.Author.User = %+v", branches[0].Target.Author.User)
	}
	if branches[1].Target.Author.User != nil {
		t.Errorf("branches[1].Target.Author.User = %+v, want nil", branches[1].Target.Author.User)
	}
}
