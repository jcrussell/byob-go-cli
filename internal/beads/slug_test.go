package beads

import "testing"

func TestSlug(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello World", "hello-world"},
		{"Foo, Bar! Baz?", "foo-bar-baz"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"already-hyphenated", "already-hyphenated"},
		{"UPPER and lower 123", "upper-and-lower-123"},
		{"runs --- of --- hyphens", "runs-of-hyphens"},
		{"", ""},
		{"!!!", ""},
	}
	for _, tc := range cases {
		if got := Slug(tc.in); got != tc.want {
			t.Errorf("Slug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
