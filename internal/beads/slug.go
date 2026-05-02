package beads

import (
	"regexp"
	"strings"
)

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slug returns the lowercase hyphen-collapsed form used as the on-disk
// directory name for an epic and its children. Mirrors `_slug` in
// scripts/convert.py:46.
func Slug(title string) string {
	return strings.Trim(slugNonAlnum.ReplaceAllString(strings.ToLower(title), "-"), "-")
}
