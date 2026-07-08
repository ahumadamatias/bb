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
	Path string
	Line int
	Task bool
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
		Use:   "comment <id>",
		Short: "Agrega un comentario a un pull request",
		Long: `Agrega un comentario a un pull request: general, o inline sobre una línea
específica del diff con --path (y opcionalmente --line). --task lo marca
además como task (bloqueante) — útil para filtrar comentarios que hay que
resolver antes de mergear.`,
		Example: `  $ bb pr comment 42 --body "LGTM"
  $ bb pr comment 42 --body "esto está mal" --path internal/api/client.go --line 42
  $ bb pr comment 42 --body "hay que corregir esto antes de mergear" --path internal/api/client.go --line 42 --task`,
		Args: cmdutil.ExactArgs(1, "bb pr comment requiere exactamente un argumento: el ID del pull request"),
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
	cmd.Flags().StringVar(&opts.Path, "path", "", "Ruta del archivo, para comentar una línea específica del diff (comentario inline)")
	cmd.Flags().IntVar(&opts.Line, "line", 0, "Número de línea en la versión nueva del archivo (requiere --path)")
	cmd.Flags().BoolVar(&opts.Task, "task", false, "Marcar el comentario como task (bloqueante)")

	return cmd
}

func commentRun(opts *CommentOptions) error {
	if opts.Body == "" {
		return cmdutil.FlagErrorf("--body es requerido")
	}
	if opts.Line != 0 && opts.Path == "" {
		return cmdutil.FlagErrorf("--line requiere --path")
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	comment, err := client.CreatePullRequestComment(workspace, repo, opts.ID, api.CreateCommentInput{
		Body: opts.Body,
		Path: opts.Path,
		Line: opts.Line,
	})
	if err != nil {
		return err
	}

	io := opts.IO
	fmt.Fprintf(io.Out, "%s Comentario agregado al PR #%d (id %d)\n", io.Green("✓"), opts.ID, comment.ID)

	if opts.Task {
		task, err := client.CreatePullRequestTask(workspace, repo, opts.ID, opts.Body, comment.ID)
		if err != nil {
			return fmt.Errorf("el comentario se creó (id %d) pero no se pudo marcar como task: %w", comment.ID, err)
		}
		fmt.Fprintf(io.Out, "%s Marcado como task bloqueante (id %d, estado %s)\n", io.Yellow("⚠"), task.ID, task.State)
	}

	return nil
}
