package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWalk_multipleParentlessFilesIsAnError(t *testing.T) {
	tmp := t.TempDir()
	dec := filepath.Join(tmp, "decisions")
	mem := filepath.Join(tmp, "memories")
	cat := filepath.Join(dec, "demo")
	if err := os.MkdirAll(cat, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(mem, 0o755); err != nil {
		t.Fatal(err)
	}
	// Two parentless files in the same directory: classic data bug.
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(cat, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("demo-epic.md",
		"---\nid: demo-epic\ntitle: Demo Epic\ntype: epic\npriority: 2\nstatus: open\nlabels: []\n---\n\n## Description\n\nFirst orphan.\n")
	write("demo-stray.md",
		"---\nid: demo-stray\ntitle: Demo Stray\ntype: decision\npriority: 2\nstatus: open\nlabels: []\n---\n\n## Description\n\nSecond orphan.\n")

	_, err := Walk(dec, mem)
	if err == nil {
		t.Fatal("expected error on category with two parentless files; got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "multiple parentless files") {
		t.Errorf("error should mention 'multiple parentless files', got: %v", err)
	}
	if !strings.Contains(msg, "demo-epic") || !strings.Contains(msg, "demo-stray") {
		t.Errorf("error should name both offenders, got: %v", err)
	}
}
