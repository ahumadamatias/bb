// Package gitctx infers the current Bitbucket workspace, repository, and
// branch by shelling out to the local git binary.
package gitctx

import (
	"errors"
	"net/url"
	"os/exec"
	"strings"
)

// ErrNotBitbucketRepo is returned when the current directory isn't a git
// repository, has no "origin" remote, or that remote isn't bitbucket.org.
var ErrNotBitbucketRepo = errors.New("no se pudo inferir el repositorio de Bitbucket: parado en un repo git con remote 'origin' apuntando a bitbucket.org, o usá --workspace y --repo")

// Context holds the workspace/repo/branch inferred from the local git repo.
type Context struct {
	Workspace string
	Repo      string
	Branch    string
}

// Current inspects the git repository in the working directory and
// returns its Bitbucket workspace, repo slug, and current branch.
func Current() (*Context, error) {
	remote, err := remoteURL("origin")
	if err != nil {
		return nil, ErrNotBitbucketRepo
	}

	workspace, repo, err := ParseRemoteURL(remote)
	if err != nil {
		return nil, err
	}

	branch, _ := currentBranch()

	return &Context{
		Workspace: workspace,
		Repo:      repo,
		Branch:    branch,
	}, nil
}

func remoteURL(remoteName string) (string, error) {
	out, err := exec.Command("git", "remote", "get-url", remoteName).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func currentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ParseRemoteURL extracts the workspace and repo slug from a bitbucket.org
// git remote URL. It supports:
//
//	git@bitbucket.org:workspace/repo.git      (SSH shorthand)
//	ssh://git@bitbucket.org/workspace/repo.git (SSH URL)
//	https://bitbucket.org/workspace/repo.git   (HTTPS)
//	https://user@bitbucket.org/workspace/repo.git (HTTPS with userinfo)
//
// The ".git" suffix is optional in all forms.
func ParseRemoteURL(raw string) (workspace, repo string, err error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, ".git")
	raw = strings.TrimSuffix(raw, "/")

	if raw == "" {
		return "", "", ErrNotBitbucketRepo
	}

	// SSH shorthand: git@bitbucket.org:workspace/repo
	if strings.HasPrefix(raw, "git@") {
		rest := strings.TrimPrefix(raw, "git@")
		host, path, ok := strings.Cut(rest, ":")
		if !ok || !isBitbucketHost(host) {
			return "", "", ErrNotBitbucketRepo
		}
		return splitWorkspaceRepo(path)
	}

	// Any URL form: ssh://, https://, http://
	if strings.Contains(raw, "://") {
		u, parseErr := url.Parse(raw)
		if parseErr != nil || !isBitbucketHost(u.Hostname()) {
			return "", "", ErrNotBitbucketRepo
		}
		return splitWorkspaceRepo(u.Path)
	}

	return "", "", ErrNotBitbucketRepo
}

func isBitbucketHost(host string) bool {
	return strings.EqualFold(host, "bitbucket.org")
}

func splitWorkspaceRepo(path string) (string, string, error) {
	path = strings.Trim(path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrNotBitbucketRepo
	}
	// repo may still contain extra path segments in malformed input; take
	// only the first segment after workspace.
	repo := parts[1]
	if idx := strings.Index(repo, "/"); idx != -1 {
		repo = repo[:idx]
	}
	return parts[0], repo, nil
}
