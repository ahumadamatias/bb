package pr

import (
	"encoding/json"
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

// ViewOptions holds the dependencies and flag values for `bb pr view`.
type ViewOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string
	Output    string

	ID   int
	Diff bool
	Web  bool
}

// NewCmdView builds the `bb pr view` command.
func NewCmdView(f *factory.Factory) *cobra.Command {
	opts := &ViewOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "Muestra un pull request",
		Example: `  $ bb pr view 42
  $ bb pr view 42 --diff
  $ bb pr view 42 --web`,
		Args: cmdutil.ExactArgs(1, "bb pr view requiere exactamente un argumento: el ID del pull request"),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("el ID del pull request debe ser un número: %q", args[0])
			}
			opts.ID = id
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			opts.Output = f.Options.Output
			return viewRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.Diff, "diff", false, "Mostrar el diff del pull request")
	cmd.Flags().BoolVar(&opts.Web, "web", false, "Abrir el pull request en el navegador (no imprime nada más)")

	return cmd
}

func viewRun(opts *ViewOptions) error {
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

	if opts.Web {
		return cmdutil.OpenInBrowser(pr.Links.HTML.Href)
	}

	var diff string
	if opts.Diff {
		diff, err = client.GetPullRequestDiff(workspace, repo, opts.ID)
		if err != nil {
			return err
		}
	}

	io := opts.IO

	if opts.Output == "json" {
		out := struct {
			*api.PullRequest
			Diff string `json:"diff,omitempty"`
		}{PullRequest: pr, Diff: diff}
		enc := json.NewEncoder(io.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	fmt.Fprintf(io.Out, "%s %s\n", io.Bold(fmt.Sprintf("#%d", pr.ID)), pr.Title)
	fmt.Fprintf(io.Out, "%s • %s → %s • %s • aprobaciones: %d\n",
		pr.State, pr.SourceBranch(), pr.DestinationBranch(), pr.Author.DisplayName, pr.ApprovalCount())
	fmt.Fprintf(io.Out, "Creado: %s\n", pr.CreatedOn.Format("2006-01-02 15:04"))
	if pr.Description != "" {
		fmt.Fprintf(io.Out, "\n%s\n", pr.Description)
	}
	if opts.Diff {
		fmt.Fprintf(io.Out, "\n%s\n", diff)
	}
	return nil
}
