package beads

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	frontmatterRE = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n?`)
	h2LineRE      = regexp.MustCompile(`^## (\S[^\n]*?)\n?$`)
)

// SplitFrontmatter parses YAML frontmatter from the head of text, returning
// the decoded map and the remaining body. Errors out if there's no
// `---\n...\n---` block at the top.
func SplitFrontmatter(text string) (map[string]any, string, error) {
	m := frontmatterRE.FindStringSubmatchIndex(text)
	if m == nil {
		return nil, "", fmt.Errorf("no frontmatter found")
	}
	var fm map[string]any
	if err := yaml.Unmarshal([]byte(text[m[2]:m[3]]), &fm); err != nil {
		return nil, "", fmt.Errorf("frontmatter yaml: %w", err)
	}
	return fm, text[m[1]:], nil
}

// SplitH2Sections slices body into {section_name_lower: content} on `## `
// lines, skipping `## ` lines that fall inside triple-backtick fenced code
// blocks (so a sample heading inside a Design code block doesn't get
// mistaken for a real section). Mirrors `_split_h2_sections` in
// scripts/convert.py:51.
func SplitH2Sections(body string) map[string]string {
	sections := map[string]string{}
	var currentName string
	currentStart := 0
	inFence := false
	pos := 0
	hasCurrent := false

	for _, line := range splitKeep(body) {
		switch {
		case strings.HasPrefix(strings.TrimLeft(line, " "), "```"):
			inFence = !inFence
		case !inFence:
			if m := h2LineRE.FindStringSubmatch(line); m != nil {
				if hasCurrent {
					sections[currentName] = strings.Trim(body[currentStart:pos], "\n")
				}
				currentName = strings.ToLower(strings.TrimSpace(m[1]))
				currentStart = pos + len(line)
				hasCurrent = true
			}
		}
		pos += len(line)
	}
	if hasCurrent {
		sections[currentName] = strings.Trim(body[currentStart:], "\n")
	}
	return sections
}

// splitKeep is `strings.SplitAfter(s, "\n")` minus the trailing empty
// element that Split* leaves when the input ends with "\n".
func splitKeep(s string) []string {
	parts := strings.SplitAfter(s, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
