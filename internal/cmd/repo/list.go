// Package repo implements `bb repo` subcommands.
package repo

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/gitctx"
	"github.com/matiasahumada/bb/internal/iostreams"
	"github.com/matiasahumada/bb/internal/tableprinter"
)

// ListOptions holds the dependencies and flag values for `bb repo list`.
type ListOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Output    string
	Limit     int
}

// NewCmdList builds the `bb repo list` command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista los repositorios de un workspace",
		Example: `  $ bb repo list
  $ bb repo list --workspace myworkspace --limit 50
  $ bb repo list --output json | jq '.[0].name'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = f.Options.Workspace
			opts.Output = f.Options.Output
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Máximo de repositorios a listar (0 = todos)")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, err := cmdutil.ResolveWorkspace(opts.GitContext, opts.Config, opts.Workspace)
	if err != nil {
		return err
	}

	repos, err := client.ListRepositories(workspace, opts.Limit)
	if err != nil {
		return err
	}

	io := opts.IO

	if opts.Output == "json" {
		enc := json.NewEncoder(io.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(repos)
	}

	t := tableprinter.New(io.Out, io.IsStdoutTTY())
	t.AddHeader("NAME", "DESCRIPTION", "UPDATED", "PRIVATE")
	for _, r := range repos {
		private := "false"
		if r.IsPrivate {
			private = "true"
		}
		t.AddRow(r.Name, cmdutil.Truncate(r.Description, 60), cmdutil.RelativeTime(r.UpdatedOn), private)
	}
	if len(repos) == 0 && io.IsStdoutTTY() {
		fmt.Fprintln(io.ErrOut, "No hay repositorios en el workspace "+strconv.Quote(workspace))
		return nil
	}
	return t.Render()
}
