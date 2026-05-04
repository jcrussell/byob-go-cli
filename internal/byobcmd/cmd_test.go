package byobcmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		name        string
		in          error
		wantFlagErr bool
	}{
		{"nil", nil, false},
		{"unknown command", fmt.Errorf(`unknown command "bogus" for "byob"`), true},
		{"already FlagError", &cmdutil.FlagError{Err: errors.New("x")}, true},
		{"plain runtime", errors.New("connection refused"), false},
	}
	for _, tc := range cases {
		got := classify(tc.in)
		var fe *cmdutil.FlagError
		isFE := got != nil && errors.As(got, &fe)
		if isFE != tc.wantFlagErr {
			t.Errorf("%s: classify→FlagError? got=%v, want=%v (err=%v)", tc.name, isFE, tc.wantFlagErr, got)
		}
	}
}

// TestRun_exitCodes drives the runner end-to-end with a tiny cobra tree
// and asserts the exit-code mapping for the byob-errors.1 vocabulary.
func TestRun_exitCodes(t *testing.T) {
	cases := []struct {
		name string
		args []string
		runE func(*cobra.Command, []string) error
		want int
	}{
		{"success", []string{"sub"}, func(*cobra.Command, []string) error { return nil }, 0},
		{"silent → 1", []string{"sub"}, func(*cobra.Command, []string) error { return cmdutil.ErrSilent }, 1},
		{"cancel → 2", []string{"sub"}, func(*cobra.Command, []string) error { return cmdutil.ErrCancel }, 2},
		{"flag error → 2", []string{"sub", "--bogus"}, nil, 2},
		{"unknown command → 2", []string{"missing"}, nil, 2},
		{"runtime error → 1", []string{"sub"}, func(*cobra.Command, []string) error { return errors.New("oops") }, 1},
	}
	for _, tc := range cases {
		root := &cobra.Command{Use: "byob", SilenceUsage: true, SilenceErrors: true}
		root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
			return &cmdutil.FlagError{Err: err}
		})
		runE := tc.runE
		if runE == nil {
			runE = func(*cobra.Command, []string) error { return nil }
		}
		root.AddCommand(&cobra.Command{Use: "sub", RunE: runE})

		ios, _, _, _ := iostreams.Test()
		got := Run(root, tc.args, ios)
		if got != tc.want {
			t.Errorf("%s: exit=%d, want %d", tc.name, got, tc.want)
		}
	}
}
