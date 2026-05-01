---
id: byob-iostreams.3
title: Data to Out, chatter and prompts to ErrOut
type: decision
priority: 2
status: open
parent: byob-iostreams
labels:
- cli
- go
- iostreams
---

## Description

Problem: `mytool list | wc -l` breaks when `mytool list` interleaves "Loading
items…" chatter into stdout. Users then reach for `--quiet` flags, which
multiply.

Idea: reserve `Out` for the command's DATA — the thing a user might pipe into
jq, awk, grep. Chatter (status messages, progress, prompts, warnings) goes to
`ErrOut`. Scripting users get clean stdout by default; humans still see the
progress because ErrOut isn't redirected by `|`.

Rule of thumb: if removing the print would change the meaning of piped
output, it goes to `Out`. Otherwise `ErrOut`. Prompts ALWAYS go to ErrOut
because prompts read from stdin, and stdin is usually the pipe.

Tradeoffs: authors must think about which stream a message belongs on. That
thinking is exactly the value.

## Design

```go
// chatter -> ErrOut
fmt.Fprintln(io.ErrOut, "Loading items from", cfg.ServerURL)
spinner.Start(io.ErrOut)

// data -> Out
for _, item := range items {
    fmt.Fprintln(io.Out, item.Name)
}

// prompts -> ErrOut (and read from In)
fmt.Fprint(io.ErrOut, "Delete all? [y/N] ")
```

