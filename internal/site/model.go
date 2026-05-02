// Package site walks decisions/ and memories/ and renders them as a
// browsable static site under _site/. Decisions are grouped by their
// on-disk category subdirectory; memories live on a single anchored page.
package site

import (
	"html/template"
	"sort"
)

// Site is the root data passed to templates.
type Site struct {
	Categories []*Category
	Memories   []*Memory

	BaseURL   string        // e.g. "/byob-go-cli" or "" for local preview
	RepoURL   string        // e.g. "https://github.com/jcrussell/byob-go-cli"
	IntroHTML template.HTML // homepage intro, rendered from README.md

	idToPath map[string]string
}

// Category groups one epic with its ordered children.
type Category struct {
	Slug     string // directory name (lowercase-hyphen)
	Epic     *Decision
	Children []*Decision
}

// Title returns the human-readable name for the category, falling back
// to the slug when the epic is missing (which would be a data bug, not a
// runtime case we expect).
func (c *Category) Title() string {
	if c.Epic != nil {
		return c.Epic.Title
	}
	return c.Slug
}

// Decision is one rendered decision or epic page.
type Decision struct {
	ID       string
	Title    string
	Type     string // "decision" or "epic"
	ParentID string
	Labels   []string

	// raw markdown captured during walk; rendered to HTML in render.go
	RawDescription string
	RawDesign      string

	DescriptionHTML template.HTML
	DesignHTML      template.HTML

	Path     string // route, e.g. "/storage/byob-storage.1/" (no BaseURL)
	Category *Category
	Prev     *Decision
	Next     *Decision
}

// IsEpic reports whether this Decision is the category's epic page.
func (d *Decision) IsEpic() bool { return d.Type == "epic" }

// Memory is one memory entry; HTML is the rendered body.
type Memory struct {
	Key  string
	HTML template.HTML

	raw string // markdown captured during walk; rendered in render.go
}

// sortCategories puts categories in alphabetical order by slug for stable
// nav rendering.
func sortCategories(cs []*Category) {
	sort.Slice(cs, func(i, j int) bool { return cs[i].Slug < cs[j].Slug })
}

// sortChildren puts decision IDs in numeric ascending order
// (byob-foo.1, .2, .10) rather than lexicographic.
func sortChildren(ds []*Decision) {
	sort.Slice(ds, func(i, j int) bool { return idLess(ds[i].ID, ds[j].ID) })
}

// idLess compares two byob-* ids by their numeric suffix when present,
// falling back to lexicographic for non-numeric tails.
func idLess(a, b string) bool {
	an, aok := splitID(a)
	bn, bok := splitID(b)
	if aok && bok && an.prefix == bn.prefix {
		return an.num < bn.num
	}
	return a < b
}

type idParts struct {
	prefix string
	num    int
}

func splitID(id string) (idParts, bool) {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '.' {
			n := 0
			ok := i+1 < len(id)
			for j := i + 1; j < len(id); j++ {
				if id[j] < '0' || id[j] > '9' {
					ok = false
					break
				}
				n = n*10 + int(id[j]-'0')
			}
			if ok {
				return idParts{prefix: id[:i], num: n}, true
			}
			return idParts{}, false
		}
	}
	return idParts{}, false
}
