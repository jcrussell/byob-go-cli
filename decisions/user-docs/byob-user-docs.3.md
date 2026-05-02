---
id: byob-user-docs.3
title: Invest in `--help` Long descriptions and Example blocks
type: decision
priority: 2
status: open
parent: byob-user-docs
labels:
  - cli
  - go
  - user-docs
---

## Description

Problem: treating `--help` as "the flag list" wastes it. Cobra's
`Short`, `Long`, and `Example` fields feed both terminal help output
*and* the generated reference docs (byob-output.3) — investing in them
once produces docs that reach users through every channel (TTY help,
man pages, Markdown site, shell completions' description text). The
alternative, writing flag lists in the README and leaving `--help`
minimal, gives users worse docs through every channel.

Idea: per-command discipline on three cobra fields.

- **`Short`** — one line, under ~60 chars, imperative voice ("list
  widgets", not "lists widgets" or "widget lister"). This is what
  `mytool --help` shows next to each subcommand name.
- **`Long`** — multi-paragraph prose explaining what the command
  does, when to use it vs. alternatives, and any non-obvious
  behavior. This is the one place to explain "why" — users reading
  `mytool foo --help` already know the command exists; they want the
  conceptual frame.
- **`Example`** — runnable command strings separated by blank lines,
  each preceded by a one-line comment describing the intent. Cobra
  indents and renders them in an "Examples:" block in `--help` and
  carries them verbatim into the generated reference.

`ErrHint` messages (byob-errors.2) point at narrative troubleshooting
docs (byob-user-docs.5) for non-obvious failure modes — the union of
`--help` and the troubleshooting page covers every user-facing
documentation need.

Tradeoffs: writing good Long descriptions takes real effort. It pays
back because it's the one piece of documentation that definitely
reaches the user — they ran `--help`, they're looking at it now.

## Design

```go
// pkg/cmd/widgets/list.go
func NewCmdList(f *Factory, runF func(*Options) error) *cobra.Command {
    opts := &Options{IO: f.IOStreams}
    cmd := &cobra.Command{
        Use:   "list",
        Short: "list widgets visible to the current user",
        Long: `List widgets in the default workspace.

By default, only widgets you own are shown. Pass --all to include
widgets shared with you, or --workspace to scope to a specific one.

Results are paginated; use --limit to change the page size (default
30). For scripting, --json emits the full result as a newline-delimited
stream — see mytool widgets --help for filter/template flags shared
across widget subcommands.`,
        Example: heredoc.Doc(`
            # list your widgets
            $ mytool widgets list

            # include shared widgets
            $ mytool widgets list --all

            # scripting: emit JSON
            $ mytool widgets list --json | jq '.[] | .id'
        `),
        RunE: func(c *cobra.Command, args []string) error {
            // ...
        },
    }
    return cmd
}
```

