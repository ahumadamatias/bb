package pr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/gitctx"
	"github.com/matiasahumada/bb/internal/iostreams"
)

func TestCreateRunUsesGitBranchAndRepoMainbranch(t *testing.T) {
	var gotPRRequest map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/myworkspace/myrepo":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mainbranch": map[string]string{"name": "main"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/repositories/myworkspace/myrepo/pullrequests":
			json.NewDecoder(r.Body).Decode(&gotPRRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 99,
				"links": map[string]interface{}{
					"html": map[string]string{"href": "https://bitbucket.org/myworkspace/myrepo/pull-requests/99"},
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClient("user@example.com", "token", "test")
	client.BaseURL = server.URL

	ios, _, stdout, _ := iostreams.Test()

	opts := &CreateOptions{
		IO:     ios,
		Client: func() (*api.Client, error) { return client, nil },
		GitContext: func() (*gitctx.Context, error) {
			return &gitctx.Context{Workspace: "myworkspace", Repo: "myrepo", Branch: "feature/x"}, nil
		},
		Config: func() (*config.Resolved, error) { return &config.Resolved{}, nil },
		Title:  "My PR title",
	}

	if err := createRun(opts); err != nil {
		t.Fatalf("createRun() error = %v", err)
	}

	if gotPRRequest["title"] != "My PR title" {
		t.Errorf("request title = %v, want %q", gotPRRequest["title"], "My PR title")
	}

	source := gotPRRequest["source"].(map[string]interface{})["branch"].(map[string]interface{})["name"]
	if source != "feature/x" {
		t.Errorf("source branch = %v, want feature/x (inferred from git context)", source)
	}

	dest := gotPRRequest["destination"].(map[string]interface{})["branch"].(map[string]interface{})["name"]
	if dest != "main" {
		t.Errorf("dest branch = %v, want main (inferred from repo mainbranch)", dest)
	}

	if !strings.Contains(stdout.String(), "pull-requests/99") {
		t.Errorf("stdout = %q, want it to contain the created PR URL", stdout.String())
	}
}

func TestCreateRunRequiresTitleWithoutTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	opts := &CreateOptions{
		IO:     ios,
		Client: func() (*api.Client, error) { return api.NewClient("u", "t", "test"), nil },
		GitContext: func() (*gitctx.Context, error) {
			return &gitctx.Context{Workspace: "ws", Repo: "repo", Branch: "main"}, nil
		},
		Config: func() (*config.Resolved, error) { return &config.Resolved{}, nil },
	}

	err := createRun(opts)
	if err == nil {
		t.Fatal("expected error when --title is missing and stdin isn't a TTY")
	}
}

func TestCreateRunResolvesReviewerUUIDPassthrough(t *testing.T) {
	var gotPRRequest map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			json.NewDecoder(r.Body).Decode(&gotPRRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    1,
				"links": map[string]interface{}{"html": map[string]string{"href": "https://bitbucket.org/ws/repo/pull-requests/1"}},
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	client := api.NewClient("user@example.com", "token", "test")
	client.BaseURL = server.URL

	ios, _, _, _ := iostreams.Test()

	opts := &CreateOptions{
		IO:     ios,
		Client: func() (*api.Client, error) { return client, nil },
		GitContext: func() (*gitctx.Context, error) {
			return &gitctx.Context{Workspace: "ws", Repo: "repo", Branch: "feature/x"}, nil
		},
		Config:    func() (*config.Resolved, error) { return &config.Resolved{}, nil },
		Title:     "Title",
		Dest:      "main",
		Reviewers: []string{"{22222222-2222-2222-2222-222222222222}"},
	}

	if err := createRun(opts); err != nil {
		t.Fatalf("createRun() error = %v", err)
	}

	reviewers := gotPRRequest["reviewers"].([]interface{})
	if len(reviewers) != 1 {
		t.Fatalf("reviewers = %v, want 1 entry", reviewers)
	}
	uuid := reviewers[0].(map[string]interface{})["uuid"]
	if uuid != "{22222222-2222-2222-2222-222222222222}" {
		t.Errorf("reviewer uuid = %v", uuid)
	}
}
