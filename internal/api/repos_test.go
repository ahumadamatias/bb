package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func serveFixture(t *testing.T, path string) http.HandlerFunc {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", path, err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func TestListRepositories(t *testing.T) {
	server := httptest.NewServer(serveFixture(t, "testdata/repositories.json"))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	repos, err := c.ListRepositories("myworkspace", 0)
	if err != nil {
		t.Fatalf("ListRepositories() error = %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("len(repos) = %d, want 2", len(repos))
	}
	if repos[0].Slug != "bb" || repos[0].Mainbranch.Name != "main" {
		t.Errorf("repos[0] = %+v", repos[0])
	}
	if !repos[0].IsPrivate {
		t.Errorf("repos[0].IsPrivate = false, want true")
	}
}

func TestGetRepository(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		serveFixture(t, "testdata/repository.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	repo, err := c.GetRepository("myworkspace", "bb")
	if err != nil {
		t.Fatalf("GetRepository() error = %v", err)
	}
	if want := "/repositories/myworkspace/bb"; gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
	if repo.Mainbranch.Name != "main" {
		t.Errorf("repo.Mainbranch.Name = %q, want main", repo.Mainbranch.Name)
	}
}
