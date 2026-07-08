package pr

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/gitctx"
	"github.com/matiasahumada/bb/internal/iostreams"
)

// mergeStrategyFlags maps the CLI's hyphenated flag values to the
// underscored values the Bitbucket API expects.
var mergeStrategyFlags = map[string]string{
	"merge-commit": api.MergeStrategyMergeCommit,
	"squash":       api.MergeStrategySquash,
	"fast-forward": api.MergeStrategyFastForward,
}

// MergeOptions holds the dependencies and flag values for `bb pr merge`.
type MergeOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string

	ID                int
	Strategy          string
	CloseSourceBranch bool
}

// NewCmdMerge builds the `bb pr merge` command.
func NewCmdMerge(f *factory.Factory) *cobra.Command {
	opts := &MergeOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "merge <id>",
		Short: "Mergea un pull request",
		Example: `  $ bb pr merge 42
  $ bb pr merge 42 --strategy squash --close-source-branch`,
		Args: cmdutil.ExactArgs(1, "bb pr merge requiere exactamente un argumento: el ID del pull request"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("el ID del pull request debe ser un número: %q", args[0])
			}
			opts.ID = id
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			return mergeRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Strategy, "strategy", "merge-commit", "Estrategia de merge: merge-commit, squash, fast-forward")
	cmd.Flags().BoolVar(&opts.CloseSourceBranch, "close-source-branch", false, "Cerrar la rama origen al mergear")

	return cmd
}

func mergeRun(opts *MergeOptions) error {
	apiStrategy, ok := mergeStrategyFlags[opts.Strategy]
	if !ok {
		return cmdutil.FlagErrorf("--strategy inválida: %q (valores soportados: merge-commit, squash, fast-forward)", opts.Strategy)
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	pr, err := client.GetPullRequest(workspace, repo, opts.ID)
	if err != nil {
		return err
	}
	if pr.State != api.PullRequestStateOpen {
		return fmt.Errorf("el pull request #%d no está abierto (estado actual: %s)", opts.ID, pr.State)
	}

	merged, err := client.MergePullRequest(workspace, repo, opts.ID, api.MergePullRequestInput{
		MergeStrategy:     apiStrategy,
		CloseSourceBranch: opts.CloseSourceBranch,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "%s Pull request #%d mergeado (%s)\n", opts.IO.Green("✓"), merged.ID, merged.State)
	return nil
}
