// Package cmdutil holds small helpers shared across bb's cobra command
// packages: flag/usage error wrapping and argument validators that map
// to exit code 2 (usage error), as opposed to exit code 1 (general
// error) or 4 (authentication error).
package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// FlagError marks an error as a usage error (bad flags/args), so main.go
// can map it to exit code 2 instead of the general exit code 1.
type FlagError struct {
	Err error
}

func (e *FlagError) Error() string { return e.Err.Error() }
func (e *FlagError) Unwrap() error { return e.Err }

// FlagErrorf builds a *FlagError from a format string.
func FlagErrorf(format string, args ...interface{}) error {
	return &FlagError{Err: fmt.Errorf(format, args...)}
}

// ExactArgs returns a cobra.PositionalArgs validator requiring exactly n
// args, producing a *FlagError (rather than cobra's default error) on
// mismatch so it maps to exit code 2.
func ExactArgs(n int, usage string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != n {
			return FlagErrorf("%s", usage)
		}
		return nil
	}
}
