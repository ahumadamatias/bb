// Package factory builds the real (non-test) dependencies used by every
// bb subcommand: config resolution, an authenticated API client, git
// context detection, and IOStreams. Commands never construct these
// themselves; they receive a *Factory and store its funcs on their
// Options struct.
package factory

import (
	"github.com/ahumadamatias/bb/internal/api"
	"github.com/ahumadamatias/bb/internal/config"
	"github.com/ahumadamatias/bb/internal/gitctx"
	"github.com/ahumadamatias/bb/internal/iostreams"
)

// GlobalOptions carries the values of bb's persistent (global) flags.
// root.go binds these via pflag.StringVar before Execute() runs, so by
// the time any RunE (and therefore any Factory func) executes, they
// hold the user's actual flag values.
type GlobalOptions struct {
	Email     string
	Token     string
	Workspace string
	Repo      string
	Output    string
}

// Factory wires together the dependencies every subcommand needs.
// Fields are funcs (not concrete values) so fake factories in tests can
// substitute their own behavior without touching real config, network,
// or git state.
type Factory struct {
	IOStreams *iostreams.IOStreams

	Config     func() (*config.Resolved, error)
	Client     func() (*api.Client, error)
	GitContext func() (*gitctx.Context, error)

	// Options points at the same struct root.go binds its persistent
	// flags to, so commands can read the final --workspace/--repo/
	// --output values inside RunE (after cobra has parsed flags).
	Options *GlobalOptions

	Version string
}

// New builds the real Factory used by cmd/bb/main.go.
func New(version string, opts *GlobalOptions) *Factory {
	ios := iostreams.System()

	configFn := func() (*config.Resolved, error) {
		return config.Resolve(config.ResolveOptions{
			Email:     opts.Email,
			Token:     opts.Token,
			Workspace: opts.Workspace,
		})
	}

	clientFn := func() (*api.Client, error) {
		resolved, err := configFn()
		if err != nil {
			return nil, err
		}
		if resolved.Email == "" || resolved.Token == "" {
			return nil, config.ErrNoCredentials
		}
		return api.NewClient(resolved.Email, resolved.Token, version), nil
	}

	gitContextFn := func() (*gitctx.Context, error) {
		return gitctx.Current()
	}

	return &Factory{
		IOStreams:  ios,
		Config:     configFn,
		Client:     clientFn,
		GitContext: gitContextFn,
		Options:    opts,
		Version:    version,
	}
}
