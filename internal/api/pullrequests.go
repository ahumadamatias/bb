package api

import (
	"fmt"
	"net/http"
	"time"
)

// Merge strategies accepted by POST .../pullrequests/{id}/merge.
const (
	MergeStrategyMergeCommit = "merge_commit"
	MergeStrategySquash      = "squash"
	MergeStrategyFastForward = "fast_forward"
)

// Pull request states as returned by the API and accepted by the
// ?state= query parameter on the list endpoint.
const (
	PullRequestStateOpen     = "OPEN"
	PullRequestStateMerged   = "MERGED"
	PullRequestStateDeclined = "DECLINED"
)

// branchRef matches the {"branch": {"name": "..."}} shape used for both
// "source" and "destination" on a pull request.
type branchRef struct {
	Branch struct {
		Name string `json:"name"`
	} `json:"branch"`
}

// Participant is a reviewer or other participant on a pull request,
// including whether they've approved it.
type Participant struct {
	User     Account `json:"user"`
	Role     string  `json:"role"`
	Approved bool    `json:"approved"`
}

// PullRequest is a subset of the pull request resource returned by the
// pullrequests endpoints.
type PullRequest struct {
	ID                int           `json:"id"`
	Title             string        `json:"title"`
	Description       string        `json:"description"`
	State             string        `json:"state"`
	Author            Account       `json:"author"`
	Source            branchRef     `json:"source"`
	Destination       branchRef     `json:"destination"`
	CreatedOn         time.Time     `json:"created_on"`
	UpdatedOn         time.Time     `json:"updated_on"`
	Reviewers         []Account     `json:"reviewers"`
	Participants      []Participant `json:"participants"`
	CloseSourceBranch bool          `json:"close_source_branch"`
	Links             Links         `json:"links"`
}

// SourceBranch and DestinationBranch expose the branch names without
// callers needing to know the nested {"branch":{"name":...}} shape.
func (pr *PullRequest) SourceBranch() string      { return pr.Source.Branch.Name }
func (pr *PullRequest) DestinationBranch() string { return pr.Destination.Branch.Name }

// ApprovalCount returns how many participants have approved the PR.
func (pr *PullRequest) ApprovalCount() int {
	n := 0
	for _, p := range pr.Participants {
		if p.Approved {
			n++
		}
	}
	return n
}

// CreatePullRequestInput is the request body for creating a pull request.
type CreatePullRequestInput struct {
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	Source            branchRef `json:"source"`
	Destination       branchRef `json:"destination"`
	CloseSourceBranch bool      `json:"close_source_branch"`
	Reviewers         []Account `json:"reviewers,omitempty"`
}

// NewCreatePullRequestInput builds a CreatePullRequestInput from plain
// branch names and reviewer UUIDs.
func NewCreatePullRequestInput(title, description, sourceBranch, destBranch string, reviewerUUIDs []string, closeSourceBranch bool) CreatePullRequestInput {
	in := CreatePullRequestInput{
		Title:             title,
		Description:       description,
		CloseSourceBranch: closeSourceBranch,
	}
	in.Source.Branch.Name = sourceBranch
	in.Destination.Branch.Name = destBranch
	for _, uuid := range reviewerUUIDs {
		in.Reviewers = append(in.Reviewers, Account{UUID: uuid})
	}
	return in
}

func pullRequestsPath(workspace, repoSlug string) string {
	return repoPath(workspace, repoSlug) + "/pullrequests"
}

func pullRequestPath(workspace, repoSlug string, id int) string {
	return fmt.Sprintf("%s/%d", pullRequestsPath(workspace, repoSlug), id)
}

// ListPullRequests lists pull requests in a repository filtered by
// state (OPEN, MERGED, DECLINED). limit <= 0 fetches every page.
func (c *Client) ListPullRequests(workspace, repoSlug, state string, limit int) ([]PullRequest, error) {
	path := pullRequestsPath(workspace, repoSlug) + "?state=" + state
	return getPaginated[PullRequest](c, path, limit)
}

// GetPullRequest fetches a single pull request by ID.
func (c *Client) GetPullRequest(workspace, repoSlug string, id int) (*PullRequest, error) {
	var pr PullRequest
	if err := c.doJSON(http.MethodGet, pullRequestPath(workspace, repoSlug, id), nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// GetPullRequestDiff fetches the raw unified diff for a pull request.
func (c *Client) GetPullRequestDiff(workspace, repoSlug string, id int) (string, error) {
	return c.doRaw(http.MethodGet, pullRequestPath(workspace, repoSlug, id)+"/diff")
}

// CreatePullRequest opens a new pull request.
func (c *Client) CreatePullRequest(workspace, repoSlug string, in CreatePullRequestInput) (*PullRequest, error) {
	var pr PullRequest
	if err := c.doJSON(http.MethodPost, pullRequestsPath(workspace, repoSlug), in, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// InlineComment locates a comment on a specific file/line of the diff.
// Not used by any v1 command yet, but modeled now so inline comments can
// be added later by just wiring up flags.
type InlineComment struct {
	Path string `json:"path"`
	To   int    `json:"to,omitempty"`
}

// commentContent matches the {"raw": "..."} shape Bitbucket uses for
// comment bodies.
type commentContent struct {
	Raw string `json:"raw"`
}

// Comment is a pull request comment, general or inline.
type Comment struct {
	ID      int            `json:"id,omitempty"`
	Content commentContent `json:"content"`
	Inline  *InlineComment `json:"inline,omitempty"`
}

// CreatePullRequestComment adds a general (non-inline) comment to a pull
// request.
func (c *Client) CreatePullRequestComment(workspace, repoSlug string, id int, body string) (*Comment, error) {
	in := Comment{Content: commentContent{Raw: body}}
	var out Comment
	if err := c.doJSON(http.MethodPost, pullRequestPath(workspace, repoSlug, id)+"/comments", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ApprovePullRequest approves a pull request as the authenticated user.
func (c *Client) ApprovePullRequest(workspace, repoSlug string, id int) error {
	return c.doJSON(http.MethodPost, pullRequestPath(workspace, repoSlug, id)+"/approve", nil, nil)
}

// UnapprovePullRequest removes the authenticated user's approval.
func (c *Client) UnapprovePullRequest(workspace, repoSlug string, id int) error {
	return c.doJSON(http.MethodDelete, pullRequestPath(workspace, repoSlug, id)+"/approve", nil, nil)
}

// MergePullRequestInput is the request body for merging a pull request.
type MergePullRequestInput struct {
	MergeStrategy     string `json:"merge_strategy"`
	CloseSourceBranch bool   `json:"close_source_branch"`
}

// MergePullRequest merges a pull request using the given strategy.
func (c *Client) MergePullRequest(workspace, repoSlug string, id int, in MergePullRequestInput) (*PullRequest, error) {
	var pr PullRequest
	if err := c.doJSON(http.MethodPost, pullRequestPath(workspace, repoSlug, id)+"/merge", in, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}
