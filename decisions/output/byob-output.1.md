---
id: byob-output.1
title: 'TTY-adaptive table printer: one API, two render paths'
type: decision
priority: 2
status: open
parent: byob-output
labels:
  - output
---

## Description

Problem: a `list` command that prints a pretty ANSI table breaks every
pipeline. A `list` command that prints tab-separated values is unreadable in
a terminal. Users want both, and `--output=plain` is a clumsy workaround.

Idea: one table printer, two render paths. On `Render()`, branch on
`IsStdoutTTY()`. TTY path: colored headers, auto-width columns, fuzzy
timestamps ("2h ago"). Non-TTY path: tab-separated columns, RFC3339
timestamps, no color. Users never pass a flag; the tool does the right thing
based on where stdout points.

Tradeoffs: authors must remember timestamps, widths, and colors diverge
between paths. Push the divergence into the printer; commands stay simple.

## Design

```go
type Table struct {
    io      *iostreams.IOStreams
    headers []string
    rows    [][]string
}

func (t *Table) AddRow(cells ...string) { t.rows = append(t.rows, cells) }

func (t *Table) Render() error {
    if t.io.IsStdoutTTY() {
        return t.renderTTY()   // padded, colored, fuzzy time
    }
    return t.renderTSV()       // tab-separated, RFC3339
}

func (t *Table) renderTSV() error {
    for _, r := range t.rows {
        fmt.Fprintln(t.io.Out, strings.Join(r, "\t"))
    }
    return nil
}
```

