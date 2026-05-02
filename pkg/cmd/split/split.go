// Package split implements `byob split`: read a bd JSONL stream from
// stdin and write decisions/<cat>/<id>.md plus memories/<key>.md.
// Mirrors `cmd_split` in scripts/convert.py:187.
package split

import (
	"bufio"
	"encoding/json"
	"fmt"
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

func NewCmdSplit(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{IO: f.IOStreams}
	cmd := &cobra.Command{
		Use:   "split",
		Short: "Read a bd JSONL stream from stdin into decisions/ + memories/",
		Long: "Wipes decisions/ and memories/*.md, then writes one markdown " +
			"file per decision/epic record and per memory record. Reads JSONL " +
			"from stdin: `bd export | byob split`.",
		RunE: func(c *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return splitRun(opts)
		},
	}
	cmd.Flags().StringVar(&opts.DecisionsDir, "decisions-dir", "decisions", "decisions tree root")
	cmd.Flags().StringVar(&opts.MemoriesDir, "memories-dir", "memories", "memories directory")
	return cmd
}

func splitRun(opts *Options) error {
	if err := os.RemoveAll(opts.DecisionsDir); err != nil {
		return fmt.Errorf("wipe decisions: %w", err)
	}
	if err := os.MkdirAll(opts.DecisionsDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(opts.MemoriesDir, 0o755); err != nil {
		return err
	}
	old, _ := filepath.Glob(filepath.Join(opts.MemoriesDir, "*.md"))
	for _, p := range old {
		if err := os.Remove(p); err != nil {
			return err
		}
	}

	var (
		beadsList []*beads.Bead
		memories  []*beads.Memory
		skipped   = map[string]int{}
	)

	sc := bufio.NewScanner(opts.IO.In)
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var head struct {
			Type      string `json:"_type"`
			ID        string `json:"id"`
			IssueType string `json:"issue_type"`
		}
		if err := json.Unmarshal([]byte(line), &head); err != nil {
			return fmt.Errorf("invalid jsonl: %w", err)
		}
		if head.Type == "memory" {
			var m beads.Memory
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				return err
			}
			if m.Key != "" {
				memories = append(memories, &m)
			}
			continue
		}
		if head.ID == "" {
			continue
		}
		if head.IssueType != "decision" && head.IssueType != "epic" {
			key := head.IssueType
			if key == "" {
				key = "<missing>"
			}
			skipped[key]++
			continue
		}
		var b beads.Bead
		if err := json.Unmarshal([]byte(line), &b); err != nil {
			return err
		}
		if b.Dependencies == nil {
			b.Dependencies = []beads.Dependency{}
		}
		beadsList = append(beadsList, &b)
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	titles := make(map[string]string, len(beadsList))
	for _, b := range beadsList {
		titles[b.ID] = b.Title
	}

	for _, b := range beadsList {
		slugSource := b.Title
		if pid := b.ParentID(); pid != "" {
			if t, ok := titles[pid]; ok {
				slugSource = t
			}
		}
		slug := beads.Slug(slugSource)
		if slug == "" {
			slug = b.ID
		}
		dir := filepath.Join(opts.DecisionsDir, slug)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		md, err := b.ToMarkdown()
		if err != nil {
			return fmt.Errorf("%s: %w", b.ID, err)
		}
		if err := os.WriteFile(filepath.Join(dir, b.ID+".md"), []byte(md), 0o644); err != nil {
			return err
		}
	}

	for _, m := range memories {
		md, err := beads.MemoryToMarkdown(m)
		if err != nil {
			return fmt.Errorf("memory %s: %w", m.Key, err)
		}
		if err := os.WriteFile(filepath.Join(opts.MemoriesDir, m.Key+".md"), []byte(md), 0o644); err != nil {
			return err
		}
	}

	fmt.Fprintf(opts.IO.ErrOut, "Wrote %d decision files and %d memory files\n", len(beadsList), len(memories))
	if len(skipped) > 0 {
		keys := make([]string, 0, len(skipped))
		for k := range skipped {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		total := 0
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%d", k, skipped[k]))
			total += skipped[k]
		}
		fmt.Fprintf(opts.IO.ErrOut, "Skipped %d non-decision issues (%s)\n", total, strings.Join(parts, ", "))
	}
	return nil
}
