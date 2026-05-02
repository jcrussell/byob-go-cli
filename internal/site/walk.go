package site

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jcrussell/byob-go-cli/internal/beads"
)

// Walk reads decisionsDir and memoriesDir from disk and returns a Site
// with categories, decisions, and memories populated. HTML rendering
// happens in a separate Render pass so the walk stays purely IO.
func Walk(decisionsDir, memoriesDir string) (*Site, error) {
	s := &Site{idToPath: map[string]string{}}

	entries, err := os.ReadDir(decisionsDir)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", decisionsDir, err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cat, err := loadCategory(filepath.Join(decisionsDir, e.Name()), e.Name())
		if err != nil {
			return nil, err
		}
		s.Categories = append(s.Categories, cat)
	}
	sortCategories(s.Categories)

	for _, c := range s.Categories {
		if c.Epic != nil {
			c.Epic.Category = c
			c.Epic.Path = "/" + c.Slug + "/"
			s.idToPath[c.Epic.ID] = c.Epic.Path
		}
		for i, d := range c.Children {
			d.Category = c
			d.Path = "/" + c.Slug + "/" + d.ID + "/"
			s.idToPath[d.ID] = d.Path
			if i > 0 {
				d.Prev = c.Children[i-1]
			}
			if i < len(c.Children)-1 {
				d.Next = c.Children[i+1]
			}
		}
	}

	for _, c := range s.Categories {
		if c.Epic == nil {
			return nil, fmt.Errorf("category %s has no epic file", c.Slug)
		}
		for _, d := range c.Children {
			if d.ParentID != c.Epic.ID {
				return nil, fmt.Errorf("decision %s has parent %q but lives under %s (epic %s)",
					d.ID, d.ParentID, c.Slug, c.Epic.ID)
			}
		}
	}

	mem, err := filepath.Glob(filepath.Join(memoriesDir, "*.md"))
	if err != nil {
		return nil, err
	}
	for _, p := range mem {
		text, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		m, err := beads.MemoryFromMarkdown(string(text))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		// HTML rendering deferred to render.go; carry the raw value here.
		s.Memories = append(s.Memories, &Memory{Key: m.Key, raw: m.Value})
	}

	return s, nil
}

func loadCategory(dir, slug string) (*Category, error) {
	matches, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	cat := &Category{Slug: slug}
	for _, p := range matches {
		text, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		b, err := beads.FromMarkdown(string(text))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		d := &Decision{
			ID:             b.ID,
			Title:          b.Title,
			Type:           b.IssueType,
			ParentID:       b.ParentID(),
			Labels:         b.Labels,
			RawDescription: b.Description,
			RawDesign:      b.Design,
		}
		// A category's "epic" is the parentless file in the directory —
		// usually type=epic, but a few one-off categories (e.g.
		// agent-onboarding) have a single parentless decision instead.
		if d.ParentID == "" {
			cat.Epic = d
		} else {
			cat.Children = append(cat.Children, d)
		}
	}
	sortChildren(cat.Children)
	return cat, nil
}
