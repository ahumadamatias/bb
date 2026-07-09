package pr

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ahumadamatias/bb/internal/api"
	"github.com/ahumadamatias/bb/internal/cmd/cmdutil"
	"github.com/ahumadamatias/bb/internal/cmd/factory"
	"github.com/ahumadamatias/bb/internal/config"
	"github.com/ahumadamatias/bb/internal/gitctx"
	"github.com/ahumadamatias/bb/internal/iostreams"
)

// ApproveOptions holds the dependencies and flag values for `bb pr approve`.
type ApproveOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string

	ID     int
	Remove bool
}

// NewCmdApprove builds the `bb pr approve` command.
func NewCmdApprove(f *factory.Factory) *cobra.Command {
	opts := &ApproveOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "approve <id>",
		Short: "Aprueba un pull request",
		Example: `  $ bb pr approve 42
  $ bb pr approve 42 --remove`,
		Args: cmdutil.ExactArgs(1, "bb pr approve requiere exactamente un argumento: el ID del pull request"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("el ID del pull request debe ser un número: %q", args[0])
			}
			opts.ID = id
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			return approveRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Remove, "remove", false, "Retirar la aprobación en lugar de aprobar")

	return cmd
}

func approveRun(opts *ApproveOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	if opts.Remove {
		if err := client.UnapprovePullRequest(workspace, repo, opts.ID); err != nil {
			return err
		}
		fmt.Fprintf(opts.IO.Out, "%s Aprobación retirada del PR #%d\n", opts.IO.Green("✓"), opts.ID)
		return nil
	}

	if err := client.ApprovePullRequest(workspace, repo, opts.ID); err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.Out, "%s PR #%d aprobado\n", opts.IO.Green("✓"), opts.ID)
	return nil
}
