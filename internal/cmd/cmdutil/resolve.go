package cmdutil

import (
	"github.com/ahumadamatias/bb/internal/config"
	"github.com/ahumadamatias/bb/internal/gitctx"
)

// ResolveWorkspace determines which workspace a command should operate
// on, for commands that don't need a specific repo (e.g. `bb repo
// list`). Precedence: --workspace flag/BB_WORKSPACE env (already merged
// into workspaceFlag by root.go) > git remote inference > config file's
// default_workspace.
func ResolveWorkspace(gitContext func() (*gitctx.Context, error), cfg func() (*config.Resolved, error), workspaceFlag string) (string, error) {
	if workspaceFlag != "" {
		return workspaceFlag, nil
	}

	if gitContext != nil {
		if gc, err := gitContext(); err == nil && gc.Workspace != "" {
			return gc.Workspace, nil
		}
	}

	resolved, err := cfg()
	if err != nil {
		return "", err
	}
	if resolved.DefaultWorkspace != "" {
		return resolved.DefaultWorkspace, nil
	}

	return "", FlagErrorf("no se pudo determinar el workspace: usá --workspace, seteá BB_WORKSPACE, o configurá default_workspace con 'bb auth login'")
}

// ResolveWorkspaceRepo determines the workspace and repo slug a command
// should operate on, for commands scoped to a single repository (branch
// list, all pr subcommands). Same precedence as ResolveWorkspace, plus
// git remote inference of the repo slug.
func ResolveWorkspaceRepo(gitContext func() (*gitctx.Context, error), cfg func() (*config.Resolved, error), workspaceFlag, repoFlag string) (workspace, repo string, err error) {
	var gc *gitctx.Context
	if gitContext != nil {
		gc, _ = gitContext()
	}

	workspace = workspaceFlag
	if workspace == "" && gc != nil {
		workspace = gc.Workspace
	}
	if workspace == "" {
		resolved, cfgErr := cfg()
		if cfgErr == nil {
			workspace = resolved.DefaultWorkspace
		}
	}

	repo = repoFlag
	if repo == "" && gc != nil {
		repo = gc.Repo
	}

	if workspace == "" || repo == "" {
		return "", "", FlagErrorf("%s", gitctx.ErrNotBitbucketRepo.Error())
	}
	return workspace, repo, nil
}
