// Package join implements `byob join`: walk decisions/ and memories/
// and emit one bd JSONL record per file to stdout. Inverse of split.
// Mirrors `cmd_join` in scripts/convert.py:261.
package join

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jcrussell/byob-go-cli/internal/beads"
	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

type Options struct {
	IO           *iostreams.IOStreams
	DecisionsDir string
	MemoriesDir  string
}

func NewCmdJoin(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{IO: f.IOStreams}
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Walk decisions/ + memories/ and emit one JSONL record per file",
		Long: "Walks decisions/**/*.md in sorted path order then memories/*.md " +
			"in sorted order, emitting one JSON record per line on stdout — " +
			"the format `bd import` consumes.",
		RunE: func(c *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return joinRun(opts)
		},
	}
	cmd.Flags().StringVar(&opts.DecisionsDir, "decisions-dir", "decisions", "decisions tree root")
	cmd.Flags().StringVar(&opts.MemoriesDir, "memories-dir", "memories", "memories directory")
	return cmd
}

func joinRun(opts *Options) error {
	enc := json.NewEncoder(opts.IO.Out)
	enc.SetEscapeHTML(false)

	var decisionPaths []string
	if _, err := os.Stat(opts.DecisionsDir); err == nil {
		err := filepath.WalkDir(opts.DecisionsDir, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !d.IsDir() && strings.HasSuffix(p, ".md") {
				decisionPaths = append(decisionPaths, p)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	sort.Strings(decisionPaths)
	for _, p := range decisionPaths {
		text, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		b, err := beads.FromMarkdown(string(text))
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if b.Dependencies == nil {
			b.Dependencies = []beads.Dependency{}
		}
		if err := enc.Encode(b); err != nil {
			return err
		}
	}

	if _, err := os.Stat(opts.MemoriesDir); err == nil {
		matches, err := filepath.Glob(filepath.Join(opts.MemoriesDir, "*.md"))
		if err != nil {
			return err
		}
		sort.Strings(matches)
		for _, p := range matches {
			text, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			m, err := beads.MemoryFromMarkdown(string(text))
			if err != nil {
				return fmt.Errorf("%s: %w", p, err)
			}
			if err := enc.Encode(m); err != nil {
				return err
			}
		}
	}
	return nil
}
