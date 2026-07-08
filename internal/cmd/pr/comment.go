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

// CommentOptions holds the dependencies and flag values for `bb pr comment`.
type CommentOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string

	ID   int
	Body string
}

// NewCmdComment builds the `bb pr comment` command.
func NewCmdComment(f *factory.Factory) *cobra.Command {
	opts := &CommentOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:     "comment <id>",
		Short:   "Agrega un comentario general a un pull request",
		Example: `  $ bb pr comment 42 --body "LGTM"`,
		Args:    cmdutil.ExactArgs(1, "bb pr comment requiere exactamente un argumento: el ID del pull request"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("el ID del pull request debe ser un número: %q", args[0])
			}
			opts.ID = id
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			return commentRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Body, "body", "", "Texto del comentario")

	return cmd
}

func commentRun(opts *CommentOptions) error {
	if opts.Body == "" {
		return cmdutil.FlagErrorf("--body es requerido")
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	comment, err := client.CreatePullRequestComment(workspace, repo, opts.ID, opts.Body)
	if err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.Out, "%s Comentario agregado al PR #%d (id %d)\n", opts.IO.Green("✓"), opts.ID, comment.ID)
	return nil
}
