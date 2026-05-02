// Package root assembles the top-level cobra command and wires every
// subcommand. Adding a verb is one import + one AddCommand line.
package root

import (
	"github.com/spf13/cobra"

	joincmd "github.com/jcrussell/byob-go-cli/pkg/cmd/join"
	splitcmd "github.com/jcrussell/byob-go-cli/pkg/cmd/split"
	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
)

func NewCmdRoot(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   f.ExecutableName,
		Short: "byob-go-cli build tooling",
		Long: "Build and tooling commands for byob-go-cli: convert beads JSONL " +
			"to and from the markdown trees under decisions/ and memories/, and " +
			"render the static site published to GitHub Pages.",

		// byob-errors.4: own all error formatting in the runner; cobra
		// shouldn't print usage on runtime errors or echo error strings.
		SilenceUsage:  true,
		SilenceErrors: true,

		// No default action: invoking `byob` with no args prints help.
		RunE: func(c *cobra.Command, args []string) error { return c.Help() },
	}
	cmd.AddCommand(splitcmd.NewCmdSplit(f, nil))
	cmd.AddCommand(joincmd.NewCmdJoin(f, nil))
	return cmd
}
