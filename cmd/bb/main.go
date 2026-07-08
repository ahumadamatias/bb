// Command bb is a CLI for Bitbucket Cloud. See internal/cmd/root for the
// command tree; this file only wires exit codes.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/matiasahumada/bb/internal/api"
	"github.com/matiasahumada/bb/internal/cmd/cmdutil"
	"github.com/matiasahumada/bb/internal/cmd/root"
	"github.com/matiasahumada/bb/internal/config"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	cmd := root.NewCmdRoot(version)

	err := cmd.Execute()
	if err == nil {
		return 0
	}

	fmt.Fprintln(os.Stderr, "Error:", err)

	var flagErr *cmdutil.FlagError
	if errors.As(err, &flagErr) {
		return 2
	}

	if errors.Is(err, config.ErrNoCredentials) {
		return 4
	}

	var apiErr *api.Error
	if errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403) {
		return 4
	}

	return 1
}
