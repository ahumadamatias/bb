package api

import (
	"net/http/httptest"
	"testing"
)

func TestCurrentUser(t *testing.T) {
	server := httptest.NewServer(serveFixture(t, "testdata/user.json"))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	u, err := c.CurrentUser()
	if err != nil {
		t.Fatalf("CurrentUser() error = %v", err)
	}
	if u.DisplayName != "Matias Ahumada" {
		t.Errorf("DisplayName = %q, want Matias Ahumada", u.DisplayName)
	}
}
