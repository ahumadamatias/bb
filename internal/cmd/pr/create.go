package pr

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/gitctx"
	"github.com/matiasahumada/bb/internal/iostreams"
)

// CreateOptions holds the dependencies and flag values for `bb pr create`.
type CreateOptions struct {
	IO         *iostreams.IOStreams
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)
	Config     func() (*config.Resolved, error)

	Workspace string
	Repo      string

	Title             string
	Body              string
	Source            string
	Dest              string
	Reviewers         []string
	CloseSourceBranch bool
	Web               bool
}

// NewCmdCreate builds the `bb pr create` command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &CreateOptions{
		IO:         f.IOStreams,
		Client:     f.Client,
		GitContext: f.GitContext,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Crea un pull request",
		Long: `Crea un pull request. --source por defecto es la rama git actual;
--dest por defecto es la rama principal del repositorio.`,
		Example: `  $ bb pr create --title "Mi cambio"
  $ bb pr create --title "Fix bug" --body "Detalle del fix" --dest main
  $ bb pr create --title "Fix bug" --reviewer ana,juan --close-source-branch --web`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = f.Options.Workspace
			opts.Repo = f.Options.Repo
			return createRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Title, "title", "", "Título del pull request")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Descripción del pull request")
	cmd.Flags().StringVar(&opts.Source, "source", "", "Rama origen (por defecto: la rama actual)")
	cmd.Flags().StringVar(&opts.Dest, "dest", "", "Rama destino (por defecto: la rama principal del repositorio)")
	cmd.Flags().StringSliceVar(&opts.Reviewers, "reviewer", nil, "Reviewer a agregar (nickname, display name o UUID); repetible o separado por comas")
	cmd.Flags().BoolVar(&opts.CloseSourceBranch, "close-source-branch", false, "Cerrar la rama origen al mergear")
	cmd.Flags().BoolVar(&opts.Web, "web", false, "Abrir el pull request creado en el navegador")

	return cmd
}

func createRun(opts *CreateOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	workspace, repo, err := cmdutil.ResolveWorkspaceRepo(opts.GitContext, opts.Config, opts.Workspace, opts.Repo)
	if err != nil {
		return err
	}

	io := opts.IO

	title := opts.Title
	if title == "" {
		if !io.IsStdinTTY() {
			return cmdutil.FlagErrorf("--title es requerido cuando no hay una terminal interactiva")
		}
		fmt.Fprint(io.Out, "Título: ")
		line, err := bufio.NewReader(io.In).ReadString('\n')
		if err != nil {
			return fmt.Errorf("leyendo título: %w", err)
		}
		title = strings.TrimSpace(line)
		if title == "" {
			return cmdutil.FlagErrorf("--title es requerido")
		}
	}

	source := opts.Source
	if source == "" {
		gc, gcErr := opts.GitContext()
		if gcErr != nil || gc.Branch == "" {
			return cmdutil.FlagErrorf("--source es requerido: no se pudo inferir la rama actual")
		}
		source = gc.Branch
	}

	dest := opts.Dest
	if dest == "" {
		repoInfo, repoErr := client.GetRepository(workspace, repo)
		if repoErr != nil {
			return fmt.Errorf("no se pudo determinar la rama principal del repositorio: %w", repoErr)
		}
		if repoInfo.Mainbranch.Name == "" {
			return cmdutil.FlagErrorf("--dest es requerido: el repositorio no tiene una rama principal configurada")
		}
		dest = repoInfo.Mainbranch.Name
	}

	reviewerUUIDs, err := resolveReviewers(client, workspace, opts.Reviewers)
	if err != nil {
		return err
	}

	in := api.NewCreatePullRequestInput(title, opts.Body, source, dest, reviewerUUIDs, opts.CloseSourceBranch)
	pr, err := client.CreatePullRequest(workspace, repo, in)
	if err != nil {
		return err
	}

	fmt.Fprintf(io.Out, "%s Pull request #%d creado: %s\n", io.Green("✓"), pr.ID, pr.Links.HTML.Href)

	if opts.Web {
		return cmdutil.OpenInBrowser(pr.Links.HTML.Href)
	}
	return nil
}

// resolveReviewers turns --reviewer values (nickname, display name, or
// already-a-UUID) into the UUIDs the API requires, fetching the
// workspace's member list only if at least one name needs resolving.
func resolveReviewers(client *api.Client, workspace string, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	needsLookup := false
	for _, n := range names {
		if !looksLikeUUID(n) {
			needsLookup = true
			break
		}
	}

	var members []api.Account
	if needsLookup {
		var err error
		members, err = client.ListWorkspaceMembers(workspace)
		if err != nil {
			return nil, fmt.Errorf("resolviendo --reviewer: %w", err)
		}
	}

	uuids := make([]string, 0, len(names))
	for _, n := range names {
		if looksLikeUUID(n) {
			uuids = append(uuids, n)
			continue
		}

		var matches []api.Account
		for _, m := range members {
			if strings.EqualFold(m.Nickname, n) || strings.EqualFold(m.DisplayName, n) || strings.EqualFold(m.Username, n) {
				matches = append(matches, m)
			}
		}

		switch len(matches) {
		case 0:
			return nil, cmdutil.FlagErrorf("--reviewer %q no coincide con ningún miembro del workspace %q", n, workspace)
		case 1:
			uuids = append(uuids, matches[0].UUID)
		default:
			names := make([]string, len(matches))
			for i, m := range matches {
				names[i] = fmt.Sprintf("%s (%s)", m.DisplayName, m.Nickname)
			}
			return nil, cmdutil.FlagErrorf("--reviewer %q es ambiguo, coincide con: %s", n, strings.Join(names, ", "))
		}
	}
	return uuids, nil
}

func looksLikeUUID(s string) bool {
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}
