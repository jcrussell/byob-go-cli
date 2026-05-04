// Package byobcmd is the middle tier between cmd/byob/main.go and the
// per-verb command packages: it owns the error→exit-code mapping and
// any process-global concerns (signal handling lands here when needed).
// See byob-layout.1.
package byobcmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

// Run executes root with args and maps the resulting error to an exit
// code. Per byob-errors.1: commands return errors, never call os.Exit.
func Run(root *cobra.Command, args []string, ios *iostreams.IOStreams) int {
	root.SetArgs(args)
	root.SetIn(ios.In)
	root.SetOut(ios.Out)
	root.SetErr(ios.ErrOut)

	err := classify(root.Execute())
	switch {
	case err == nil:
		return 0
	case errors.Is(err, cmdutil.ErrCancel):
		return 2
	case errors.Is(err, cmdutil.ErrSilent):
		return 1
	case errors.As(err, new(*cmdutil.FlagError)):
		fmt.Fprintln(ios.ErrOut, "error:", err)
		return 2
	default:
		fmt.Fprintln(ios.ErrOut, "error:", err)
		return 1
	}
}

// classify wraps cobra-emitted "unknown command" errors as FlagError so
// they exit 2 alongside flag-parse errors. Cobra has no public sentinel
// for unknown-command, so we match on the message prefix it emits in
// (*Command).findSuggestions / Execute. SetFlagErrorFunc on root handles
// the pflag side (unknown flag, missing arg).
func classify(err error) error {
	if err == nil {
		return nil
	}
	var fe *cmdutil.FlagError
	if errors.As(err, &fe) {
		return err
	}
	if strings.HasPrefix(err.Error(), "unknown command ") {
		return &cmdutil.FlagError{Err: err}
	}
	return err
}
