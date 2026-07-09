// Package branch implements `bb branch` subcommands.
package branch

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/ahumadamatias/bb/internal/api"
	"github.com/ahumadamatias/bb/internal/cmd/cmdutil"
	"github.com/ahumadamatias/bb/internal/cmd/factory"
	"github.com/ahumadamatias/bb/internal/config"
	"github.com/ahumadamatias/bb/internal/gitctx"
	"github.com/ahumadamatias/bb/internal/iostreams"
	"github.com/ahumadamatias/bb/internal/tableprinter"
)

// ListOptions holds the dependencies and flag values for `bb branch list`.
type ListOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string
	Output    string
	Limit     int
}

// NewCmdList builds the `bb branch list` command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista las ramas del repositorio actual",
		Example: `  $ bb branch list
  $ bb branch list --workspace myworkspace --repo myrepo`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			opts.Output = f.Options.Output
			return listRun(opts)
		},
	}

	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Máximo de ramas a listar (0 = todas)")

	return cmd
}

func listRun(opts *ListOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	branches, err := client.ListBranches(workspace, repo, opts.Limit)
	if err != nil {
		return err
	}

	io := opts.IO

	if opts.Output == "json" {
		enc := json.NewEncoder(io.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(branches)
	}

	t := tableprinter.New(io.Out, io.IsStdoutTTY())
	t.AddHeader("NAME", "LAST COMMIT", "AUTHOR", "DATE")
	for _, b := range branches {
		hash := b.Target.Hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		author := b.Target.Author.Raw
		if b.Target.Author.User != nil && b.Target.Author.User.DisplayName != "" {
			author = b.Target.Author.User.DisplayName
		}
		date := ""
		if !b.Target.Date.IsZero() {
			date = b.Target.Date.Format("2006-01-02")
		}
		t.AddRow(b.Name, hash, author, date)
	}
	return t.Render()
}
