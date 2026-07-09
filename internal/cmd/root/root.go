// Package root assembles bb's full command tree and global flags.
package root

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/ahumadamatias/bb/internal/cmd/auth"
	"github.com/ahumadamatias/bb/internal/cmd/branch"
	"github.com/ahumadamatias/bb/internal/cmd/cmdutil"
	"github.com/ahumadamatias/bb/internal/cmd/factory"
	"github.com/ahumadamatias/bb/internal/cmd/pr"
	"github.com/ahumadamatias/bb/internal/cmd/repo"
)

// NewCmdRoot builds the root "bb" command, wires the real Factory, and
// registers every subcommand.
func NewCmdRoot(version string) *cobra.Command {
	opts := &factory.GlobalOptions{}

	cmd := &cobra.Command{
		Use:   "bb <command> <subcommand> [flags]",
		Short: "bb es un CLI no oficial para Bitbucket Cloud",
		Long: `bb es un CLI no oficial para interactuar con Bitbucket Cloud (bitbucket.org)
desde la terminal: repositorios, ramas y pull requests, pensado tanto para
uso diario como para scripting y agentes de código.`,
		Example: `  $ bb auth login
  $ bb repo list
  $ bb pr create --title "Mi cambio" --dest main
  $ bb pr list --output json | jq '.[0].id'`,
		Version:           version,
		SilenceErrors:     true,
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return &cmdutil.FlagError{Err: err}
	})

	pf := cmd.PersistentFlags()
	pf.StringVar(&opts.Email, "email", os.Getenv("BB_EMAIL"), "Email de Atlassian (o seteá BB_EMAIL)")
	pf.StringVar(&opts.Token, "token", os.Getenv("BB_TOKEN"), "API token de Atlassian (o seteá BB_TOKEN)")
	pf.StringVarP(&opts.Workspace, "workspace", "w", os.Getenv("BB_WORKSPACE"), "Workspace de Bitbucket (o seteá BB_WORKSPACE); por defecto se infiere del remote git")
	pf.StringVar(&opts.Repo, "repo", "", "Repositorio de Bitbucket; por defecto se infiere del remote git")
	pf.StringVarP(&opts.Output, "output", "o", "", `Formato de salida: "" (texto/tablas) o "json"`)

	cmd.PersistentPreRunE = func(c *cobra.Command, args []string) error {
		if opts.Output != "" && opts.Output != "json" {
			return cmdutil.FlagErrorf("--output inválido: %q (valores soportados: json)", opts.Output)
		}
		return nil
	}

	f := factory.New(version, opts)

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Autenticate contra Bitbucket Cloud",
	}
	authCmd.AddCommand(auth.NewCmdLogin(f))
	authCmd.AddCommand(auth.NewCmdStatus(f))
	cmd.AddCommand(authCmd)

	repoCmd := &cobra.Command{
		Use:   "repo",
		Short: "Trabajá con repositorios de Bitbucket",
	}
	repoCmd.AddCommand(repo.NewCmdList(f))
	cmd.AddCommand(repoCmd)

	branchCmd := &cobra.Command{
		Use:   "branch",
		Short: "Trabajá con ramas de un repositorio",
	}
	branchCmd.AddCommand(branch.NewCmdList(f))
	cmd.AddCommand(branchCmd)

	cmd.AddCommand(pr.NewCmdPR(f))

	return cmd
}
