---
key: quote-strings-in-errors
---

Format user-supplied strings in errors and log messages with `%q` not
`%s`: `fmt.Errorf("unknown subcommand %q", name)`. `%q` wraps the
value in Go-syntax quotes and escapes non-printables — empty strings
render as `""` instead of vanishing into mid-sentence whitespace,
trailing newlines surface as `\n`, and a quoted `"--help"` doesn't
get visually merged with surrounding text.
