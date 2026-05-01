---
key: quote-strings-in-errors
---

When an error or log message contains a user-supplied string, format it
with `%q` not `%s`: `fmt.Errorf("unknown subcommand %q", name)` instead
of `... %s`. `%q` wraps the value in Go-syntax quotes and escapes
non-printables, so empty strings render as `""` (not vanishing into
mid-sentence whitespace), trailing newlines surface as `\n`, and a
quoted argument like `"--help"` doesn't get visually merged with
surrounding text. The Google Go Style decisions call this out
explicitly; the cost is one character per format verb and the payoff
is debugging time saved on every "looks empty but isn't" mystery.
