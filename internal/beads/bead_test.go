package beads

import (
	"reflect"
	"strings"
	"testing"
)

func TestBeadRoundtrip(t *testing.T) {
	in := &Bead{
		ID:        "byob-foo.1",
		Title:     "A title with \"quotes\"",
		IssueType: "byob",
		Priority:  2,
		Status:    "open",
		// ToMarkdown sorts labels; declare them sorted so the
		// roundtrip Labels comparison is symmetric.
		Labels: []string{"cli", "go"},
		// SplitH2Sections lowercases section names, so the roundtrip
		// preserves Description / Design content but not casing.
		Description: "Body para.\n\nSecond para.",
		Design:      "```\ncode\n```\n\nText.",
		Dependencies: []Dependency{
			{IssueID: "byob-foo.1", DependsOnID: "byob-foo", Type: "parent-child", Metadata: "{}"},
		},
	}
	md, err := in.ToMarkdown()
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if !strings.HasPrefix(md, "---\n") {
		t.Fatalf("expected frontmatter delimiter prefix, got %q", md[:20])
	}
	got, err := FromMarkdown(md)
	if err != nil {
		t.Fatalf("FromMarkdown: %v", err)
	}

	if got.ID != in.ID {
		t.Errorf("ID: got %q, want %q", got.ID, in.ID)
	}
	if got.Title != in.Title {
		t.Errorf("Title: got %q, want %q", got.Title, in.Title)
	}
	if got.IssueType != in.IssueType {
		t.Errorf("IssueType: got %q, want %q", got.IssueType, in.IssueType)
	}
	if got.Priority != in.Priority {
		t.Errorf("Priority: got %d, want %d", got.Priority, in.Priority)
	}
	if got.Status != in.Status {
		t.Errorf("Status: got %q, want %q", got.Status, in.Status)
	}
	if !reflect.DeepEqual(got.Labels, in.Labels) {
		t.Errorf("Labels: got %v, want %v", got.Labels, in.Labels)
	}
	if got.Description != in.Description {
		t.Errorf("Description: got %q, want %q", got.Description, in.Description)
	}
	if got.Design != in.Design {
		t.Errorf("Design: got %q, want %q", got.Design, in.Design)
	}
	if got.ParentID() != "byob-foo" {
		t.Errorf("ParentID: got %q, want %q", got.ParentID(), "byob-foo")
	}
}

func TestMemoryRoundtrip(t *testing.T) {
	in := &Memory{Type: "memory", Key: "errors-wrap-w", Value: "Wrap with %w."}
	md, err := MemoryToMarkdown(in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := MemoryFromMarkdown(md)
	if err != nil {
		t.Fatal(err)
	}
	if got.Key != in.Key || got.Value != in.Value {
		t.Errorf("got {%q, %q}, want {%q, %q}", got.Key, got.Value, in.Key, in.Value)
	}
}

func TestSplitH2Sections_skipsFencedHeading(t *testing.T) {
	body := "## Description\n\nIntro.\n\n## Design\n\n```\n## Not a section\n```\n\nAfter.\n"
	got := SplitH2Sections(body)
	if _, exists := got["not a section"]; exists {
		t.Errorf("fence-internal heading was incorrectly promoted to a section: %v", got)
	}
	if !strings.Contains(got["design"], "## Not a section") {
		t.Errorf("design section should contain the fenced fake heading verbatim, got %q", got["design"])
	}
}
