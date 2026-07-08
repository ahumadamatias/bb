package auth

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/factory"
	"github.com/matiasahumada/bb/internal/config"
	"github.com/matiasahumada/bb/internal/iostreams"
)

// LoginOptions holds the dependencies and flag values for `bb auth login`.
type LoginOptions struct {
	IO        *iostreams.IOStreams
	NewClient func(email, token string) *api.Client

	Email string
	Token string
}

// NewCmdLogin builds the `bb auth login` command.
func NewCmdLogin(f *factory.Factory) *cobra.Command {
	opts := &LoginOptions{
		IO: f.IOStreams,
		NewClient: func(email, token string) *api.Client {
			return api.NewClient(email, token, f.Version)
		},
	}

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Configura el email y API token para autenticarte con Bitbucket Cloud",
		Long: `Configura las credenciales de bb: un email de Atlassian y un API token
(creado en https://id.atlassian.com con scopes de repositorios y pull requests).

Sin --email/--token, pregunta ambos de forma interactiva (el token no se muestra
en pantalla). Valida las credenciales contra la API antes de guardarlas.`,
		Example: `  $ bb auth login
  $ bb auth login --email vos@example.com --token TU_API_TOKEN`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return loginRun(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Email, "email", "", "Email de tu cuenta de Atlassian")
	cmd.Flags().StringVar(&opts.Token, "token", "", "API token de Atlassian")

	return cmd
}

func loginRun(opts *LoginOptions) error {
	io := opts.IO

	email := opts.Email
	if email == "" {
		if !io.IsStdinTTY() {
			return cmdutil.FlagErrorf("--email es requerido cuando no hay una terminal interactiva")
		}
		fmt.Fprint(io.Out, "Email de Atlassian: ")
		line, err := bufio.NewReader(io.In).ReadString('\n')
		if err != nil {
			return fmt.Errorf("leyendo email: %w", err)
		}
		email = strings.TrimSpace(line)
	}

	token := opts.Token
	if token == "" {
		if !io.IsStdinTTY() {
			return cmdutil.FlagErrorf("--token es requerido cuando no hay una terminal interactiva")
		}
		fmt.Fprint(io.Out, "API token: ")
		t, err := io.ReadPassword()
		fmt.Fprintln(io.Out)
		if err != nil {
			return fmt.Errorf("leyendo token: %w", err)
		}
		token = strings.TrimSpace(t)
	}

	if email == "" || token == "" {
		return cmdutil.FlagErrorf("email y token son requeridos")
	}

	client := opts.NewClient(email, token)
	user, err := client.CurrentUser()
	if err != nil {
		return fmt.Errorf("no se pudieron validar las credenciales: %w", err)
	}

	cfg := &config.Config{Email: email, Token: token}

	if io.IsStdinTTY() {
		if workspaces, wsErr := client.ListWorkspaces(0); wsErr == nil && len(workspaces) > 0 {
			fmt.Fprintln(io.Out, "\nWorkspaces disponibles:")
			for i, w := range workspaces {
				fmt.Fprintf(io.Out, "  %d. %s (%s)\n", i+1, w.Slug, w.Name)
			}
			fmt.Fprint(io.Out, "Elegí un workspace por defecto (número, Enter para omitir): ")
			line, _ := bufio.NewReader(io.In).ReadString('\n')
			line = strings.TrimSpace(line)
			if idx, convErr := strconv.Atoi(line); convErr == nil && idx >= 1 && idx <= len(workspaces) {
				cfg.DefaultWorkspace = workspaces[idx-1].Slug
			}
		}
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("guardando configuración: %w", err)
	}

	fmt.Fprintf(io.Out, "%s Autenticado como %s (%s)\n", io.Green("✓"), user.DisplayName, email)
	if cfg.DefaultWorkspace != "" {
		fmt.Fprintf(io.Out, "Workspace por defecto: %s\n", cfg.DefaultWorkspace)
	}
	return nil
}
