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
// "To" anchors to a line in the new (destination) version of the file,
// which covers added/unchanged lines — the common case. Bitbucket also
// supports a "from" line for comments anchored to a removed line in the
// old version, not modeled here yet.
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

// CreateCommentInput is the input to CreatePullRequestComment. Leave
// Path empty for a general (non-inline) comment; set Path (and
// optionally Line) to anchor the comment to a specific file/line of the
// diff.
type CreateCommentInput struct {
	Body string
	Path string
	Line int
}

// CreatePullRequestComment adds a comment to a pull request: general, or
// inline on a specific file/line when in.Path is set.
func (c *Client) CreatePullRequestComment(workspace, repoSlug string, id int, in CreateCommentInput) (*Comment, error) {
	body := Comment{Content: commentContent{Raw: in.Body}}
	if in.Path != "" {
		body.Inline = &InlineComment{Path: in.Path, To: in.Line}
	}
	var out Comment
	if err := c.doJSON(http.MethodPost, pullRequestPath(workspace, repoSlug, id)+"/comments", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Task states as returned by the API in the "state" field. Bitbucket
// also has a "pending" boolean on tasks, but its polarity is
// inconsistently documented/behaves confusingly in practice — this
// client only relies on "state" and never sends "pending".
const (
	TaskStateUnresolved = "UNRESOLVED"
	TaskStateResolved   = "RESOLVED"
)

// Task is a pull request task: a to-do / blocking item, optionally
// attached to a comment (it then renders below that comment in
// Bitbucket's UI).
type Task struct {
	ID      int            `json:"id"`
	State   string         `json:"state"`
	Content commentContent `json:"content"`
}

// taskCommentRef references an existing comment by ID when creating a
// task linked to it; Bitbucket only requires the ID field for this.
type taskCommentRef struct {
	ID int `json:"id"`
}

// createTaskInput is the request body for creating a pull request task.
type createTaskInput struct {
	Content commentContent  `json:"content"`
	Comment *taskCommentRef `json:"comment,omitempty"`
}

// CreatePullRequestTask creates a task (to-do / blocking item) on a pull
// request. If commentID is non-zero, the task is linked to that
// existing comment; pass 0 for a standalone task not tied to any
// comment.
func (c *Client) CreatePullRequestTask(workspace, repoSlug string, prID int, body string, commentID int) (*Task, error) {
	in := createTaskInput{Content: commentContent{Raw: body}}
	if commentID != 0 {
		in.Comment = &taskCommentRef{ID: commentID}
	}
	var out Task
	if err := c.doJSON(http.MethodPost, pullRequestPath(workspace, repoSlug, prID)+"/tasks", in, &out); err != nil {
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
