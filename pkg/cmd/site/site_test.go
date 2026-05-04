package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	sitepkg "github.com/jcrussell/byob-go-cli/internal/site"
	"github.com/jcrussell/byob-go-cli/pkg/cmdutil"
	"github.com/jcrussell/byob-go-cli/pkg/iostreams"
)

func TestNewCmdSite_runFOverride(t *testing.T) {
	var captured *Options
	cmd := NewCmdSite(&cmdutil.Factory{IOStreams: iostreams.System()}, func(o *Options) error {
		captured = o
		return nil
	})
	cmd.SetArgs([]string{"--out", "/tmp/s", "--base-url", "/x", "--strict"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if captured.OutDir != "/tmp/s" || captured.BaseURL != "/x" || !captured.Strict {
		t.Errorf("flags not wired: %+v", captured)
	}
}

func TestSiteSmoke_writesExpectedFiles(t *testing.T) {
	tmp := t.TempDir()
	dec := filepath.Join(tmp, "decisions")
	mem := filepath.Join(tmp, "memories")
	out := filepath.Join(tmp, "_site")

	mustMkdir(t, filepath.Join(dec, "demo"))
	mustWrite(t, filepath.Join(dec, "demo", "demo-epic.md"),
		"---\nid: demo-epic\ntitle: Demo Epic\ntype: epic\npriority: 2\nstatus: open\nlabels: []\n---\n\n## Description\n\nIntro.\n")
	mustWrite(t, filepath.Join(dec, "demo", "demo.1.md"),
		"---\nid: demo.1\ntitle: First\ntype: decision\npriority: 2\nstatus: open\nparent: demo-epic\nlabels: []\n---\n\n## Description\n\nSee demo-epic for context.\n")
	mustMkdir(t, mem)
	mustWrite(t, filepath.Join(mem, "tip.md"),
		"---\nkey: tip\n---\n\nA tip.\n")

	s, err := sitepkg.Walk(dec, mem)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	s.BaseURL = "/x"
	s.RepoURL = "https://example.com/repo"

	if err := sitepkg.Render(s, out, "", "", false, nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	for _, p := range []string{
		filepath.Join(out, "index.html"),
		filepath.Join(out, "404.html"),
		filepath.Join(out, "decisions", "index.html"),
		filepath.Join(out, "decisions", "demo", "index.html"),
		filepath.Join(out, "decisions", "demo", "demo.1", "index.html"),
		filepath.Join(out, "memories", "index.html"),
		filepath.Join(out, "static", "site.css"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s: %v", p, err)
		}
	}

	// Verify B1 fix: cross-link inside the description should produce an
	// anchor in the rendered HTML, but no `<code>[demo-epic](...)</code>`
	// pattern (which is what would happen if the rewriter touched
	// inline-code spans).
	body, err := os.ReadFile(filepath.Join(out, "decisions", "demo", "demo.1", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	if !strings.Contains(html, `href="/x/decisions/demo/"`) {
		t.Errorf("expected cross-link to demo-epic, got: %s", html)
	}
	if strings.Contains(html, `<code>[`) {
		t.Errorf("rewriter leaked markdown link syntax into <code>: %s", html)
	}

	// Verify B2 fix: homepage cards must not nest <a> inside <a>.
	homepage, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if hasNestedAnchor(string(homepage)) {
		t.Error("homepage contains nested <a> elements")
	}
}

func TestSiteSmoke_strictFailsOnUnknownRef(t *testing.T) {
	tmp := t.TempDir()
	dec := filepath.Join(tmp, "decisions")
	mem := filepath.Join(tmp, "memories")
	out := filepath.Join(tmp, "_site")

	mustMkdir(t, filepath.Join(dec, "demo"))
	mustWrite(t, filepath.Join(dec, "demo", "demo-epic.md"),
		"---\nid: demo-epic\ntitle: Demo Epic\ntype: epic\npriority: 2\nstatus: open\nlabels: []\n---\n\n## Description\n\nRefs byob-mystery.99.\n")
	mustMkdir(t, mem)

	s, err := sitepkg.Walk(dec, mem)
	if err != nil {
		t.Fatal(err)
	}
	if err := sitepkg.Render(s, out, "", "", true, nil); err == nil {
		t.Fatal("expected strict Render to fail on unknown cross-reference")
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, content string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(p))
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// hasNestedAnchor reports whether s contains an <a> element whose body
// contains another <a> before the matching </a>. Goldmark + our templates
// emit well-formed open tags as `<a `, `<a>`, or `<a\n`, so a simple
// depth counter on those tokens is sufficient.
func hasNestedAnchor(s string) bool {
	depth := 0
	i := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "<a ") || strings.HasPrefix(s[i:], "<a>") || strings.HasPrefix(s[i:], "<a\n") {
			if depth > 0 {
				return true
			}
			depth++
			i += 2
			continue
		}
		if strings.HasPrefix(s[i:], "</a>") {
			if depth > 0 {
				depth--
			}
			i += 4
			continue
		}
		i++
	}
	return false
}
