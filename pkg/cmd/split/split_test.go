package split

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

func TestNewCmdSplit_runFOverride(t *testing.T) {
	var captured *Options
	cmd := NewCmdSplit(&cmdutil.Factory{IOStreams: iostreams.System()}, func(o *Options) error {
		captured = o
		return nil
	})
	cmd.SetArgs([]string{"--decisions-dir", "/tmp/d", "--memories-dir", "/tmp/m"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if captured == nil {
		t.Fatal("runF was not invoked")
	}
	if captured.DecisionsDir != "/tmp/d" || captured.MemoriesDir != "/tmp/m" {
		t.Errorf("flags not wired: %+v", captured)
	}
}

func TestSplitRun_writesDecisionAndMemoryFiles(t *testing.T) {
	tmp := t.TempDir()
	jsonl := strings.Join([]string{
		`{"id":"x","title":"X Cat","issue_type":"epic","priority":2,"status":"open","dependencies":[]}`,
		`{"id":"x.1","title":"First Child","issue_type":"decision","priority":2,"status":"open","dependencies":[{"issue_id":"x.1","depends_on_id":"x","type":"parent-child","metadata":"{}"}]}`,
		`{"_type":"memory","key":"errors-wrap-w","value":"Wrap with %w."}`,
		`{"id":"task.1","title":"Some Task","issue_type":"task"}`,
		"",
	}, "\n")

	ios, _, _, errBuf := iostreams.Test()
	ios.In = bytes.NewBufferString(jsonl)

	opts := &Options{
		IO:           ios,
		DecisionsDir: filepath.Join(tmp, "decisions"),
		MemoriesDir:  filepath.Join(tmp, "memories"),
	}
	if err := splitRun(opts); err != nil {
		t.Fatalf("splitRun: %v", err)
	}

	// Both x and x.1 land in the same slug dir, derived from the parent's title.
	for _, p := range []string{
		filepath.Join(tmp, "decisions", "x-cat", "x.md"),
		filepath.Join(tmp, "decisions", "x-cat", "x.1.md"),
		filepath.Join(tmp, "memories", "errors-wrap-w.md"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s: %v", p, err)
		}
	}
	// Tasks (non-decision/non-epic) must be skipped, not written.
	if matches, _ := filepath.Glob(filepath.Join(tmp, "decisions", "*", "task*.md")); len(matches) != 0 {
		t.Errorf("task issue should have been skipped, got %v", matches)
	}
	// The skip diagnostic should mention task=1.
	if !strings.Contains(errBuf.String(), "task=1") {
		t.Errorf("expected skip diagnostic mentioning task=1, got %q", errBuf.String())
	}
}

func TestSplitRun_invalidJSONL(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.In = bytes.NewBufferString("not valid json\n")
	opts := &Options{
		IO:           ios,
		DecisionsDir: filepath.Join(t.TempDir(), "decisions"),
		MemoriesDir:  filepath.Join(t.TempDir(), "memories"),
	}
	if err := splitRun(opts); err == nil {
		t.Fatal("expected error on invalid JSONL")
	}
}
