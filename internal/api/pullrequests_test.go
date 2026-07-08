package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPullRequests(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		serveFixture(t, "testdata/pullrequests_list.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	prs, err := c.ListPullRequests("myworkspace", "bb", PullRequestStateOpen, 0)
	if err != nil {
		t.Fatalf("ListPullRequests() error = %v", err)
	}
	if len(prs) != 1 || prs[0].ID != 42 {
		t.Fatalf("prs = %+v", prs)
	}
	if gotQuery != "state=OPEN" {
		t.Errorf("query = %q, want state=OPEN", gotQuery)
	}
}

func TestGetPullRequest(t *testing.T) {
	server := httptest.NewServer(serveFixture(t, "testdata/pullrequest.json"))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	pr, err := c.GetPullRequest("myworkspace", "bb", 42)
	if err != nil {
		t.Fatalf("GetPullRequest() error = %v", err)
	}
	if pr.SourceBranch() != "feature/login" || pr.DestinationBranch() != "main" {
		t.Errorf("pr source/dest = %q/%q", pr.SourceBranch(), pr.DestinationBranch())
	}
	if pr.ApprovalCount() != 1 {
		t.Errorf("ApprovalCount() = %d, want 1", pr.ApprovalCount())
	}
}

func TestGetPullRequestDiff(t *testing.T) {
	const diffText = "diff --git a/foo b/foo\n+bar\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repositories/myworkspace/bb/pullrequests/42/diff" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(diffText))
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	diff, err := c.GetPullRequestDiff("myworkspace", "bb", 42)
	if err != nil {
		t.Fatalf("GetPullRequestDiff() error = %v", err)
	}
	if diff != diffText {
		t.Errorf("diff = %q, want %q", diff, diffText)
	}
}

func TestCreatePullRequest(t *testing.T) {
	var gotBody CreatePullRequestInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		serveFixture(t, "testdata/pullrequest.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	in := NewCreatePullRequestInput("Add login flow", "desc", "feature/login", "main", []string{"{22222222-2222-2222-2222-222222222222}"}, true)
	pr, err := c.CreatePullRequest("myworkspace", "bb", in)
	if err != nil {
		t.Fatalf("CreatePullRequest() error = %v", err)
	}
	if pr.ID != 42 {
		t.Errorf("pr.ID = %d, want 42", pr.ID)
	}
	if gotBody.Source.Branch.Name != "feature/login" || gotBody.Destination.Branch.Name != "main" {
		t.Errorf("request body source/dest = %+v", gotBody)
	}
	if !gotBody.CloseSourceBranch {
		t.Errorf("request body CloseSourceBranch = false, want true")
	}
	if len(gotBody.Reviewers) != 1 || gotBody.Reviewers[0].UUID != "{22222222-2222-2222-2222-222222222222}" {
		t.Errorf("request body reviewers = %+v", gotBody.Reviewers)
	}
}

func TestCreatePullRequestComment(t *testing.T) {
	var gotPath, gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		serveFixture(t, "testdata/comment.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	comment, err := c.CreatePullRequestComment("myworkspace", "bb", 42, "Looks good to me.")
	if err != nil {
		t.Fatalf("CreatePullRequestComment() error = %v", err)
	}
	if comment.Content.Raw != "Looks good to me." {
		t.Errorf("comment.Content.Raw = %q", comment.Content.Raw)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if want := "/repositories/myworkspace/bb/pullrequests/42/comments"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestApproveAndUnapprovePullRequest(t *testing.T) {
	var gotMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if want := "/repositories/myworkspace/bb/pullrequests/42/approve"; r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	if err := c.ApprovePullRequest("myworkspace", "bb", 42); err != nil {
		t.Fatalf("ApprovePullRequest() error = %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}

	if err := c.UnapprovePullRequest("myworkspace", "bb", 42); err != nil {
		t.Fatalf("UnapprovePullRequest() error = %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
}

func TestMergePullRequest(t *testing.T) {
	var gotBody MergePullRequestInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		serveFixture(t, "testdata/pullrequest.json")(w, r)
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	_, err := c.MergePullRequest("myworkspace", "bb", 42, MergePullRequestInput{
		MergeStrategy:     MergeStrategySquash,
		CloseSourceBranch: true,
	})
	if err != nil {
		t.Fatalf("MergePullRequest() error = %v", err)
	}
	if gotBody.MergeStrategy != "squash" || !gotBody.CloseSourceBranch {
		t.Errorf("request body = %+v", gotBody)
	}
}

func TestMergePullRequestConflictError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "This pull request has conflicts and cannot be merged."},
		})
	}))
	defer server.Close()

	c := NewClient("user@example.com", "token", "test")
	c.BaseURL = server.URL

	_, err := c.MergePullRequest("myworkspace", "bb", 42, MergePullRequestInput{MergeStrategy: MergeStrategyMergeCommit})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T, want *Error", err)
	}
	if apiErr.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d, want 409", apiErr.StatusCode)
	}
	if apiErr.Message != "This pull request has conflicts and cannot be merged." {
		t.Errorf("Message = %q", apiErr.Message)
	}
}
