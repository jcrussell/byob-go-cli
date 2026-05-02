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
	"strings"

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

// Render renders s into outDir. README is the path to README.md (used for
// the homepage intro); pass "" to skip the intro.
func Render(s *Site, outDir, readmePath string, strict bool, log io.Writer) error {
	md := newGoldmark()

	// Resolve homepage intro from README if provided.
	if readmePath != "" {
		intro, err := os.ReadFile(readmePath)
		if err != nil {
			return fmt.Errorf("read README: %w", err)
		}
		s.IntroHTML = renderMarkdown(md, readmeIntro(string(intro)))
	}

	// Render decision and memory bodies.
	for _, c := range s.Categories {
		if c.Epic != nil {
			s.renderDecision(md, c.Epic, strict, log)
		}
		for _, d := range c.Children {
			s.renderDecision(md, d, strict, log)
		}
	}
	for _, m := range s.Memories {
		body := s.rewriteRefs(m.raw, strict, log)
		m.HTML = renderMarkdown(md, body)
	}

	// Build templates.
	tpl, err := template.New("").Funcs(template.FuncMap{
		"url": func(p string) string {
			if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
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

	// Categories + decisions.
	for _, c := range s.Categories {
		if err := writePage(tpl, "category.html", filepath.Join(outDir, c.Slug, "index.html"), pageData{
			Site:     s,
			Category: c,
			Title:    c.Title() + " — byob-go-cli",
		}); err != nil {
			return err
		}
		for _, d := range c.Children {
			if err := writePage(tpl, "decision.html", filepath.Join(outDir, c.Slug, d.ID, "index.html"), pageData{
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

func (s *Site) renderDecision(md goldmark.Markdown, d *Decision, strict bool, log io.Writer) {
	if d.RawDescription != "" {
		body := s.rewriteRefs(d.RawDescription, strict, log)
		d.DescriptionHTML = renderMarkdown(md, body)
	}
	if d.RawDesign != "" {
		body := s.rewriteRefs(d.RawDesign, strict, log)
		d.DesignHTML = renderMarkdown(md, body)
	}
}

// rewriteRefs replaces bare `byob-foo` / `byob-foo.N` mentions with markdown
// links to the corresponding page, skipping fenced code blocks. Unknown ids
// are left as plain text; under strict, an unknown id returns an error via
// log (not a hard failure for v1, just diagnostic).
func (s *Site) rewriteRefs(body string, strict bool, log io.Writer) string {
	var out strings.Builder
	inFence := false
	for _, line := range splitKeep(body) {
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			out.WriteString(line)
			continue
		}
		if inFence {
			out.WriteString(line)
			continue
		}
		out.WriteString(refRE.ReplaceAllStringFunc(line, func(m string) string {
			if path, ok := s.idToPath[m]; ok {
				return fmt.Sprintf("[%s](%s%s)", m, s.BaseURL, path)
			}
			if strict && log != nil {
				fmt.Fprintf(log, "unknown ref: %s\n", m)
			}
			return m
		}))
	}
	return out.String()
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
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
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
	return fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
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
	})
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
