package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/iostreams"
)

// StatusOptions holds the dependencies for `bb auth status`.
type StatusOptions struct {
	IO     *iostreams.IOStreams
	Client func() (*api.Client, error)
	Config func() (*config.Resolved, error)
}

// NewCmdStatus builds the `bb auth status` command.
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	opts := &StatusOptions{
		IO:     f.IOStreams,
		Client: f.Client,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Valida las credenciales configuradas contra la API de Bitbucket",
		Example: `  $ bb auth status`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return statusRun(opts)
		},
	}

	return cmd
}

func statusRun(opts *StatusOptions) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	user, err := client.CurrentUser()
	if err != nil {
		return err
	}

	resolved, err := opts.Config()
	if err != nil {
		return err
	}

	io := opts.IO
	fmt.Fprintf(io.Out, "%s Autenticado en bitbucket.org como %s (%s)\n", io.Green("✓"), user.DisplayName, resolved.Email)
	if resolved.DefaultWorkspace != "" {
		fmt.Fprintf(io.Out, "Workspace por defecto: %s\n", resolved.DefaultWorkspace)
	}
	return nil
}
