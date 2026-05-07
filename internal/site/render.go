package site

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

var refRE = regexp.MustCompile(`\bbyob-[a-z][a-z0-9-]*(?:\.\d+)?\b`)

// Render renders s into outDir. readmePath is the path to README.md (used
// for the homepage intro); creditsPath is the path to CREDITS.md (rendered
// at /credits/); pass "" to skip either. With strict=true, Render returns
// an error if any byob-* cross-reference can't be resolved to a known
// page (per the --strict flag's promise in pkg/cmd/site).
func Render(s *Site, outDir, readmePath, creditsPath string, strict bool, log io.Writer) error {
	md := newGoldmark()

	var unknownRefs []string
	refLog := func(id string) {
		unknownRefs = append(unknownRefs, id)
		if log != nil {
			fmt.Fprintf(log, "unknown ref: %s\n", id)
		}
	}

	// Resolve homepage intro from README if provided.
	if readmePath != "" {
		intro, err := os.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("read README: %w", err)
		}
		s.IntroHTML = renderMarkdown(md, readmeIntro(string(intro)))
	}

	// Resolve credits body if provided. Run through rewriteRefs so any
	// byob-* mentions in CREDITS.md become links and respect --strict.
	if creditsPath != "" {
		body, err := os.ReadFile(creditsPath)
		if err != nil {
			return fmt.Errorf("read CREDITS: %w", err)
		}
		s.CreditsHTML = renderMarkdown(md, s.rewriteRefs(string(body), refLog))
	}

	// Render decision and memory bodies.
	for _, c := range s.Categories {
		if c.Epic != nil {
			s.renderDecision(md, c.Epic, refLog)
		}
		for _, d := range c.Children {
			s.renderDecision(md, d, refLog)
		}
	}
	for _, m := range s.Memories {
		body := s.rewriteRefs(m.raw, refLog)
		m.HTML = renderMarkdown(md, body)
	}

	if strict && len(unknownRefs) > 0 {
		return fmt.Errorf("strict: %d unknown byob-* cross-reference(s): %s",
			len(unknownRefs), strings.Join(uniqueSorted(unknownRefs), ", "))
	}

	// Build templates.
	tpl, err := template.New("").Funcs(template.FuncMap{
		"url": func(p string) string {
			if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "//") {
				return p
			}
			return s.BaseURL + p
		},
	}).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return fmt.Errorf("parse templates: %w", err)
	}

	if err := os.RemoveAll(outDir); err != nil {
		return err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// Homepage.
	if err := writePage(tpl, "index.html", filepath.Join(outDir, "index.html"), pageData{
		Site:  s,
		Title: "byob-go-cli",
	}); err != nil {
		return err
	}

	// Memories page.
	if err := writePage(tpl, "memories.html", filepath.Join(outDir, "memories", "index.html"), pageData{
		Site:  s,
		Title: "Memories — byob-go-cli",
	}); err != nil {
		return err
	}

	// Decisions index page.
	if err := writePage(tpl, "decisions.html", filepath.Join(outDir, "decisions", "index.html"), pageData{
		Site:  s,
		Title: "Decisions — byob-go-cli",
	}); err != nil {
		return err
	}

	// Credits page (only when CREDITS.md was supplied).
	if s.CreditsHTML != "" {
		if err := writePage(tpl, "credits.html", filepath.Join(outDir, "credits", "index.html"), pageData{
			Site:  s,
			Title: "Credits — byob-go-cli",
		}); err != nil {
			return err
		}
	}

	// Categories + decisions.
	for _, c := range s.Categories {
		if err := writePage(tpl, "category.html", filepath.Join(outDir, "decisions", c.Slug, "index.html"), pageData{
			Site:     s,
			Category: c,
			Title:    c.Title() + " — byob-go-cli",
		}); err != nil {
			return err
		}
		for _, d := range c.Children {
			if err := writePage(tpl, "decision.html", filepath.Join(outDir, "decisions", c.Slug, d.ID, "index.html"), pageData{
				Site:     s,
				Category: c,
				Decision: d,
				Title:    d.Title + " — byob-go-cli",
			}); err != nil {
				return err
			}
		}
	}

	// 404 page.
	if err := writePage(tpl, "404.html", filepath.Join(outDir, "404.html"), pageData{
		Site:  s,
		Title: "Not found — byob-go-cli",
	}); err != nil {
		return err
	}

	// Static assets.
	if err := copyStatic(outDir); err != nil {
		return err
	}

	return nil
}

func (s *Site) renderDecision(md goldmark.Markdown, d *Decision, refLog func(string)) {
	if d.RawDescription != "" {
		body := s.rewriteRefs(d.RawDescription, refLog)
		d.DescriptionHTML = renderMarkdown(md, body)
	}
	if d.RawDesign != "" {
		body := s.rewriteRefs(d.RawDesign, refLog)
		d.DesignHTML = renderMarkdown(md, body)
	}
}

// rewriteRefs replaces bare `byob-foo` / `byob-foo.N` mentions with markdown
// links to the corresponding page, skipping content inside fenced code
// blocks, indented (4-space / tab) code blocks, and inline backtick spans.
// Unknown ids are reported via refLog and left as plain text.
func (s *Site) rewriteRefs(body string, refLog func(string)) string {
	var out strings.Builder
	inFence := false
	for _, line := range splitKeep(body) {
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			inFence = !inFence
			out.WriteString(line)
			continue
		}
		if inFence || isIndentedCodeBlock(line) {
			out.WriteString(line)
			continue
		}
		out.WriteString(s.rewriteOutsideCodeSpans(line, refLog))
	}
	return out.String()
}

// rewriteOutsideCodeSpans walks a single non-code line, alternating between
// inline-code regions (delimited by matched backtick runs of equal length,
// per CommonMark) and free text. The regex pass runs only on the free-text
// chunks, so `byob-foo` inside inline code is preserved verbatim.
func (s *Site) rewriteOutsideCodeSpans(line string, refLog func(string)) string {
	var out strings.Builder
	i := 0
	for i < len(line) {
		if line[i] == '`' {
			j := i
			for j < len(line) && line[j] == '`' {
				j++
			}
			tickLen := j - i
			closeOff := findCloseTicks(line[j:], tickLen)
			if closeOff < 0 {
				// Unbalanced run — CommonMark renders the rest as text;
				// our rewriter follows suit so the literal `byob-` mention
				// in prose still gets linkified.
				s.writeRewritten(&out, line[i:], refLog)
				return out.String()
			}
			end := j + closeOff + tickLen
			out.WriteString(line[i:end])
			i = end
			continue
		}
		j := i
		for j < len(line) && line[j] != '`' {
			j++
		}
		s.writeRewritten(&out, line[i:j], refLog)
		i = j
	}
	return out.String()
}

func (s *Site) writeRewritten(out *strings.Builder, text string, refLog func(string)) {
	out.WriteString(refRE.ReplaceAllStringFunc(text, func(m string) string {
		if path, ok := s.idToPath[m]; ok {
			return fmt.Sprintf("[%s](%s%s)", m, s.BaseURL, path)
		}
		// Only versioned children (`byob-foo.N`) count as "known unknowns"
		// for diagnostics. A bare `byob-foo` mention that doesn't resolve
		// is almost always prose (project names, generic byob-* terms),
		// not a stale ref worth failing strict on.
		if refLog != nil && strings.ContainsRune(m, '.') {
			refLog(m)
		}
		return m
	}))
}

// isIndentedCodeBlock reports whether line begins with the 4-space or tab
// indent that turns it into a CommonMark indented code block. The check is
// intentionally conservative: list-item continuation paragraphs at 4
// spaces look the same and will also be skipped, which is fine — missing
// a rewrite is a strictly better failure mode than mangling code samples.
func isIndentedCodeBlock(line string) bool {
	if strings.HasPrefix(line, "\t") {
		return true
	}
	return len(line) >= 4 && line[:4] == "    "
}

// findCloseTicks returns the offset of the next backtick run of exactly n
// ticks within s, or -1 if none.
func findCloseTicks(s string, n int) int {
	for i := 0; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		j := i
		for j < len(s) && s[j] == '`' {
			j++
		}
		if j-i == n {
			return i
		}
		i = j
	}
	return -1
}

func uniqueSorted(xs []string) []string {
	if len(xs) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(xs))
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}

// readmeIntro returns the README content from the start through (but not
// including) the first `## Quickstart` heading.
func readmeIntro(text string) string {
	const marker = "## Quickstart"
	if idx := strings.Index(text, "\n"+marker); idx >= 0 {
		return text[:idx]
	}
	return text
}

func newGoldmark() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
}

// chromaCSS returns a stylesheet for code blocks that pairs the "github"
// (light) and "github-dark" themes. Each theme is wrapped in its own
// `prefers-color-scheme` media query so they're mutually exclusive — this
// is the only safe way to mix two chroma styles in one stylesheet, since
// rules outside any media query would otherwise leak through to the
// non-matching mode (e.g. light's `.chroma .nx { color:#1f2328 }`
// painting package names dark-on-dark in dark mode).
func chromaCSS() ([]byte, error) {
	f := chromahtml.New()
	var buf bytes.Buffer
	buf.WriteString("/* generated from chroma styles: github + github-dark */\n")
	for _, t := range []struct {
		scheme, style string
	}{
		{"light", "github"},
		{"dark", "github-dark"},
	} {
		var inner bytes.Buffer
		if err := f.WriteCSS(&inner, styles.Get(t.style)); err != nil {
			return nil, fmt.Errorf("write %s chroma CSS: %w", t.scheme, err)
		}
		fmt.Fprintf(&buf, "\n@media (prefers-color-scheme: %s) {\n  ", t.scheme)
		buf.WriteString(strings.ReplaceAll(strings.TrimRight(inner.String(), "\n"), "\n", "\n  "))
		buf.WriteString("\n}\n")
	}
	return buf.Bytes(), nil
}

func renderMarkdown(md goldmark.Markdown, src string) template.HTML {
	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(buf.String())
}

type pageData struct {
	Site     *Site
	Category *Category
	Decision *Decision
	Title    string
}

func writePage(tpl *template.Template, tmpl, outPath string, data pageData) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, tmpl, data); err != nil {
		return fmt.Errorf("execute %s: %w", tmpl, err)
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func copyStatic(outDir string) error {
	if err := fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := staticFS.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(outDir, p) // p includes "static/" prefix
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	}); err != nil {
		return err
	}
	css, err := chromaCSS()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "static", "chroma.css"), css, 0o644)
}

// splitKeep mirrors the helper in beads/parse.go (kept private here to
// avoid a cross-package dep on an unexported helper).
func splitKeep(s string) []string {
	parts := strings.SplitAfter(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
