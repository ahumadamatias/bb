// Package pr implements `bb pr` subcommands.
package pr

import (
	"github.com/spf13/cobra"

	"github.com/matiasahumada/bb/internal/cmd/factory"
)

// NewCmdPR builds the `bb pr` command group.
func NewCmdPR(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Trabajá con pull requests",
	}

	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdView(f))
	cmd.AddCommand(NewCmdComment(f))
	cmd.AddCommand(NewCmdApprove(f))
	cmd.AddCommand(NewCmdMerge(f))

	return cmd
}
