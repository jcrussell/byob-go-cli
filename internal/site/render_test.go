package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newRefSite() *Site {
	return &Site{
		BaseURL: "/x",
		idToPath: map[string]string{
			"byob-foo.1": "/foo/byob-foo.1/",
			"byob-bar":   "/bar/",
		},
	}
}

// captureUnknowns returns a refLog that records the ids it was called with.
func captureUnknowns() (func(string), *[]string) {
	var got []string
	return func(id string) { got = append(got, id) }, &got
}

func TestRewriteRefs_bareIDIsLinkified(t *testing.T) {
	s := newRefSite()
	got := s.rewriteRefs("see byob-foo.1 here\n", nil)
	want := "see [byob-foo.1](/x/foo/byob-foo.1/) here\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRewriteRefs_inlineCodeIsPreserved(t *testing.T) {
	s := newRefSite()
	in := "see `byob-foo.1` and byob-foo.1 outside\n"
	got := s.rewriteRefs(in, nil)
	want := "see `byob-foo.1` and [byob-foo.1](/x/foo/byob-foo.1/) outside\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestRewriteRefs_doubleBacktickCodeSpan(t *testing.T) {
	s := newRefSite()
	// CommonMark allows `` ` `` (backtick inside double-backticks). Our
	// closer must match the same run length.
	in := "lit ``byob-foo.1 with ` tick`` then byob-foo.1\n"
	got := s.rewriteRefs(in, nil)
	want := "lit ``byob-foo.1 with ` tick`` then [byob-foo.1](/x/foo/byob-foo.1/)\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestRewriteRefs_fencedBlockIsPreserved(t *testing.T) {
	s := newRefSite()
	in := "before byob-foo.1\n```\nbyob-foo.1 inside\n```\nafter byob-foo.1\n"
	got := s.rewriteRefs(in, nil)
	if !strings.Contains(got, "byob-foo.1 inside") {
		t.Errorf("fenced content was rewritten; got %q", got)
	}
	if !strings.Contains(got, "before [byob-foo.1](") || !strings.Contains(got, "after [byob-foo.1](") {
		t.Errorf("non-fenced content was not rewritten; got %q", got)
	}
}

func TestRewriteRefs_indentedBlockIsPreserved(t *testing.T) {
	s := newRefSite()
	in := "Prose byob-foo.1 here.\n\n    byob-foo.1 in indented block\n\nback to byob-foo.1.\n"
	got := s.rewriteRefs(in, nil)
	if !strings.Contains(got, "    byob-foo.1 in indented block") {
		t.Errorf("indented-block content was rewritten; got %q", got)
	}
	if !strings.Contains(got, "Prose [byob-foo.1](") || !strings.Contains(got, "back to [byob-foo.1](") {
		t.Errorf("prose was not rewritten; got %q", got)
	}
}

func TestRewriteRefs_tabIndentIsPreserved(t *testing.T) {
	s := newRefSite()
	in := "Prose byob-foo.1.\n\n\tbyob-foo.1 in tab block\n"
	got := s.rewriteRefs(in, nil)
	if !strings.Contains(got, "\tbyob-foo.1 in tab block") {
		t.Errorf("tab-indented block was rewritten; got %q", got)
	}
}

func TestRewriteRefs_unknownIDIsLogged(t *testing.T) {
	s := newRefSite()
	log, captured := captureUnknowns()
	in := "see byob-mystery.7 ok\n"
	got := s.rewriteRefs(in, log)
	if got != in {
		t.Errorf("unknown id was rewritten: %q", got)
	}
	if len(*captured) != 1 || (*captured)[0] != "byob-mystery.7" {
		t.Errorf("captured = %v, want [byob-mystery.7]", *captured)
	}
}

func TestRewriteRefs_inlineCodeWithUnknownIDDoesNotLog(t *testing.T) {
	s := newRefSite()
	log, captured := captureUnknowns()
	in := "see `byob-mystery.7` ok\n"
	got := s.rewriteRefs(in, log)
	if got != in {
		t.Errorf("inline code was rewritten: %q", got)
	}
	if len(*captured) != 0 {
		t.Errorf("captured = %v, want empty (id was inside inline code)", *captured)
	}
}

func TestRewriteRefs_bareUnknownMentionIsNotLogged(t *testing.T) {
	// `byob-go-cli` (project name, no version suffix) appears in prose all
	// the time. Strict mode shouldn't flip on it — only versioned children
	// (byob-foo.N) count as "known unknowns" worth diagnosing.
	s := newRefSite()
	log, captured := captureUnknowns()
	got := s.rewriteRefs("see byob-go-cli, byob-other elsewhere\n", log)
	if strings.Contains(got, "[byob-go-cli]") || strings.Contains(got, "[byob-other]") {
		t.Errorf("bare prose mentions should not be linkified: %q", got)
	}
	if len(*captured) != 0 {
		t.Errorf("captured = %v, want empty (unversioned mentions are prose)", *captured)
	}
}

func TestRewriteRefs_unbalancedBacktickFallsThrough(t *testing.T) {
	s := newRefSite()
	in := "trailing ` byob-foo.1\n"
	got := s.rewriteRefs(in, nil)
	want := "trailing ` [byob-foo.1](/x/foo/byob-foo.1/)\n"
	if got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
}

func TestRender_strictFailsOnUnknown(t *testing.T) {
	tmp := t.TempDir()
	s := newRefSite()
	s.Categories = []*Category{{
		Slug: "demo",
		Epic: &Decision{
			ID:             "byob-demo",
			Title:          "Demo",
			Type:           "epic",
			Path:           "/demo/",
			RawDescription: "Refs byob-mystery.99.",
		},
	}}
	s.Categories[0].Epic.Category = s.Categories[0]

	err := Render(s, tmp, "", "", true, nil)
	if err == nil {
		t.Fatal("Render with strict and unknown ref returned nil; want error")
	}
	if !strings.Contains(err.Error(), "byob-mystery.99") {
		t.Errorf("error should mention the unknown id, got: %v", err)
	}
}

func TestRender_codeBlocksUseClasses(t *testing.T) {
	tmp := t.TempDir()
	s := newRefSite()
	s.Categories = []*Category{{
		Slug: "demo",
		Epic: &Decision{
			ID:        "byob-demo",
			Title:     "Demo",
			Type:      "epic",
			Path:      "/demo/",
			RawDesign: "```go\npackage main\n\nfunc main() { println(\"hi\") }\n```\n",
		},
	}}
	s.Categories[0].Epic.Category = s.Categories[0]

	if err := Render(s, tmp, "", "", false, nil); err != nil {
		t.Fatalf("Render: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(tmp, "decisions", "demo", "index.html"))
	if err != nil {
		t.Fatalf("read rendered page: %v", err)
	}
	got := string(body)
	if !strings.Contains(got, `class="chroma"`) {
		t.Errorf("rendered HTML missing class=\"chroma\"; got:\n%s", got)
	}
	if strings.Contains(got, `style="background-color`) {
		t.Errorf("rendered HTML still emits inline background-color; want class-based output")
	}
	if !strings.Contains(got, `chroma.css`) {
		t.Errorf("rendered HTML missing chroma.css stylesheet link; got:\n%s", got)
	}

	css, err := os.ReadFile(filepath.Join(tmp, "static", "chroma.css"))
	if err != nil {
		t.Fatalf("read chroma.css: %v", err)
	}
	cssText := string(css)
	if !strings.Contains(cssText, ".chroma") {
		t.Errorf("chroma.css missing .chroma rules; got:\n%s", cssText)
	}
	if !strings.Contains(cssText, "prefers-color-scheme: dark") {
		t.Errorf("chroma.css missing dark-mode media query; got:\n%s", cssText)
	}
}

func TestRender_nonStrictTolerantOfUnknown(t *testing.T) {
	tmp := t.TempDir()
	s := newRefSite()
	s.Categories = []*Category{{
		Slug: "demo",
		Epic: &Decision{
			ID:             "byob-demo",
			Title:          "Demo",
			Type:           "epic",
			Path:           "/demo/",
			RawDescription: "Refs byob-mystery.99.",
		},
	}}
	s.Categories[0].Epic.Category = s.Categories[0]

	if err := Render(s, tmp, "", "", false, nil); err != nil {
		t.Fatalf("non-strict Render unexpectedly failed: %v", err)
	}
}
