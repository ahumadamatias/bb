package pr

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/gitctx"
	"github.com/matiasahumada/bb/internal/iostreams"
	"github.com/matiasahumada/bb/internal/tableprinter"
)

// ListOptions holds the dependencies and flag values for `bb pr list`.
type ListOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string
	Output    string

	State string
	Limit int
}

// NewCmdList builds the `bb pr list` command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista los pull requests del repositorio",
		Example: `  $ bb pr list
  $ bb pr list --state MERGED
  $ bb pr list --output json | jq '.[0].id'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			opts.Output = f.Options.Output
			return listRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.State, "state", api.PullRequestStateOpen, "Filtrar por estado: OPEN, MERGED, DECLINED")
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Máximo de pull requests a listar (0 = todos)")

	return cmd
}

func listRun(opts *ListOptions) error {
	state := strings.ToUpper(opts.State)
	switch state {
	case api.PullRequestStateOpen, api.PullRequestStateMerged, api.PullRequestStateDeclined:
	default:
		return cmdutil.FlagErrorf("--state inválido: %q (valores soportados: OPEN, MERGED, DECLINED)", opts.State)
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	prs, err := client.ListPullRequests(workspace, repo, state, opts.Limit)
	if err != nil {
		return err
	}

	io := opts.IO

	if opts.Output == "json" {
		enc := json.NewEncoder(io.Out)
		enc.SetIndent("", "  ")
		return enc.Encode(prs)
	}

	t := tableprinter.New(io.Out, io.IsStdoutTTY())
	t.AddHeader("ID", "TITLE", "SOURCE → DEST", "AUTHOR", "STATE")
	for _, p := range prs {
		t.AddRow(
			strconv.Itoa(p.ID),
			cmdutil.Truncate(p.Title, 50),
			p.SourceBranch()+" → "+p.DestinationBranch(),
			p.Author.DisplayName,
			p.State,
		)
	}
	return t.Render()
}
