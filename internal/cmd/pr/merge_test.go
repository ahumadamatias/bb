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

func TestMergeRunRejectsNonOpenPR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 5, "state": "MERGED"})
	}))
	defer server.Close()

	client := api.NewClient("user@example.com", "token", "test")
	client.BaseURL = server.URL

	ios, _, _, _ := iostreams.Test()

	opts := &MergeOptions{
		IO:         ios,
		Client:     func() (*api.Client, error) { return client, nil },
		GitContext: func() (*gitctx.Context, error) { return &gitctx.Context{Workspace: "ws", Repo: "repo"}, nil },
		Config:     func() (*config.Resolved, error) { return &config.Resolved{}, nil },
		ID:         5,
		Strategy:   "merge-commit",
	}

	err := mergeRun(opts)
	if err == nil {
		t.Fatal("expected error when merging a non-OPEN pull request")
	}
	if !strings.Contains(err.Error(), "no está abierto") {
		t.Errorf("error = %v, want it to mention the PR isn't open", err)
	}
}

func TestMergeRunSuccess(t *testing.T) {
	var gotMergeBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 5, "state": "OPEN"})
		case http.MethodPost:
			json.NewDecoder(r.Body).Decode(&gotMergeBody)
			json.NewEncoder(w).Encode(map[string]interface{}{"id": 5, "state": "MERGED"})
		}
	}))
	defer server.Close()

	client := api.NewClient("user@example.com", "token", "test")
	client.BaseURL = server.URL

	ios, _, stdout, _ := iostreams.Test()

	opts := &MergeOptions{
		IO:                ios,
		Client:            func() (*api.Client, error) { return client, nil },
		GitContext:        func() (*gitctx.Context, error) { return &gitctx.Context{Workspace: "ws", Repo: "repo"}, nil },
		Config:            func() (*config.Resolved, error) { return &config.Resolved{}, nil },
		ID:                5,
		Strategy:          "squash",
		CloseSourceBranch: true,
	}

	if err := mergeRun(opts); err != nil {
		t.Fatalf("mergeRun() error = %v", err)
	}

	if gotMergeBody["merge_strategy"] != "squash" {
		t.Errorf("merge_strategy = %v, want squash", gotMergeBody["merge_strategy"])
	}
	if gotMergeBody["close_source_branch"] != true {
		t.Errorf("close_source_branch = %v, want true", gotMergeBody["close_source_branch"])
	}
	if !strings.Contains(stdout.String(), "mergeado") {
		t.Errorf("stdout = %q, want confirmation message", stdout.String())
	}
}

func TestMergeRunInvalidStrategy(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	opts := &MergeOptions{IO: ios, Strategy: "bogus"}

	err := mergeRun(opts)
	if err == nil {
		t.Fatal("expected error for an unsupported --strategy value")
	}
}
