// Package beads converts between bd's JSONL export format and the per-bead
// markdown files under decisions/ and memories/. Mirrors scripts/convert.py
// — see that file for the original reference implementation.
package beads

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Dependency is the bd parent-child link encoded inside a bead record.
type Dependency struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	Metadata    string `json:"metadata"`
}

// Bead is the JSONL shape of a decision or epic record from `bd export`.
// Personal metadata (owner, created_at, created_by, updated_at) is
// intentionally omitted — Dolt inside .beads/ keeps the real history.
type Bead struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	IssueType       string       `json:"issue_type"`
	Priority        int          `json:"priority"`
	Status          string       `json:"status"`
	Labels          []string     `json:"labels"`
	Description     string       `json:"description"`
	Design          string       `json:"design"`
	Dependencies    []Dependency `json:"dependencies"`
	DependencyCount int          `json:"dependency_count"`
	DependentCount  int          `json:"dependent_count"`
	CommentCount    int          `json:"comment_count"`
}

// Memory is the JSONL shape of a memory record. The bd export emits these
// with `_type: "memory"`; Bead/Memory share no fields, so we keep them as
// distinct types and dispatch on the `_type` discriminator at parse time.
type Memory struct {
	Type     string `json:"_type"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Category string `json:"category,omitempty"`
}

// ParentID returns the parent-child target if any, else "".
func (b *Bead) ParentID() string {
	for _, d := range b.Dependencies {
		if d.Type == "parent-child" && d.DependsOnID != "" {
			return d.DependsOnID
		}
	}
	return ""
}

// frontmatter is what we serialize at the top of each decision file.
// Field order is significant — yaml.Marshal honors struct order — and
// mirrors the order in scripts/convert.py:bead_to_md.
type frontmatter struct {
	ID       string   `yaml:"id"`
	Title    string   `yaml:"title"`
	Type     string   `yaml:"type"`
	Priority int      `yaml:"priority"`
	Status   string   `yaml:"status"`
	Parent   string   `yaml:"parent,omitempty"`
	Labels   []string `yaml:"labels"`
}

type memoryFrontmatter struct {
	Key      string `yaml:"key"`
	Category string `yaml:"category,omitempty"`
}

// ToMarkdown serializes one bead as a markdown file with YAML frontmatter
// followed by `## Description` and `## Design` sections. Mirrors
// `bead_to_md` in scripts/convert.py:81.
func (b *Bead) ToMarkdown() (string, error) {
	labels := append([]string(nil), b.Labels...)
	sort.Strings(labels)

	fm := frontmatter{
		ID:       b.ID,
		Title:    b.Title,
		Type:     valueOr(b.IssueType, "task"),
		Priority: b.Priority,
		Status:   valueOr(b.Status, "open"),
		Parent:   b.ParentID(),
		Labels:   labels,
	}

	var yb bytes.Buffer
	enc := yaml.NewEncoder(&yb)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return "", fmt.Errorf("encode frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", err
	}

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(bytes.TrimRight(yb.Bytes(), "\n"))
	out.WriteString("\n---\n\n")
	if d := strings.TrimRight(b.Description, "\n"); d != "" {
		out.WriteString("## Description\n\n")
		out.WriteString(d)
		out.WriteString("\n\n")
	}
	if d := strings.TrimRight(b.Design, "\n"); d != "" {
		out.WriteString("## Design\n\n")
		out.WriteString(d)
		out.WriteString("\n\n")
	}
	return out.String(), nil
}

// FromMarkdown is the inverse of ToMarkdown.
func FromMarkdown(text string) (*Bead, error) {
	fm, body, err := SplitFrontmatter(text)
	if err != nil {
		return nil, err
	}
	id, _ := fm["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("missing id in frontmatter")
	}
	sections := SplitH2Sections(body)
	b := &Bead{
		ID:           id,
		Title:        stringOr(fm["title"], ""),
		IssueType:    stringOr(fm["type"], "task"),
		Priority:     intOr(fm["priority"], 2),
		Status:       stringOr(fm["status"], "open"),
		Labels:       toStringSlice(fm["labels"]),
		Description:  sections["description"],
		Design:       sections["design"],
		Dependencies: []Dependency{},
	}
	if parent := stringOr(fm["parent"], ""); parent != "" {
		b.Dependencies = append(b.Dependencies, Dependency{
			IssueID:     b.ID,
			DependsOnID: parent,
			Type:        "parent-child",
			Metadata:    "{}",
		})
	}
	return b, nil
}

// MemoryToMarkdown emits a memory's frontmatter + body.
// Mirrors `memory_to_md` in scripts/convert.py:155.
func MemoryToMarkdown(m *Memory) (string, error) {
	fm := memoryFrontmatter{Key: m.Key, Category: m.Category}

	var yb bytes.Buffer
	enc := yaml.NewEncoder(&yb)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return "", fmt.Errorf("encode memory frontmatter: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", err
	}

	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(bytes.TrimRight(yb.Bytes(), "\n"))
	out.WriteString("\n---\n\n")
	out.WriteString(strings.TrimRight(m.Value, "\n"))
	out.WriteString("\n")
	return out.String(), nil
}

// MemoryFromMarkdown is the inverse of MemoryToMarkdown.
func MemoryFromMarkdown(text string) (*Memory, error) {
	fm, body, err := SplitFrontmatter(text)
	if err != nil {
		return nil, err
	}
	key, _ := fm["key"].(string)
	if key == "" {
		return nil, fmt.Errorf("missing key in memory frontmatter")
	}
	return &Memory{
		Type:     "memory",
		Key:      key,
		Value:    strings.Trim(body, "\n"),
		Category: stringOr(fm["category"], ""),
	}, nil
}

func valueOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func stringOr(v any, fallback string) string {
	s, ok := v.(string)
	if !ok || s == "" {
		return fallback
	}
	return s
}

func intOr(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return fallback
}

func toStringSlice(v any) []string {
	xs, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
