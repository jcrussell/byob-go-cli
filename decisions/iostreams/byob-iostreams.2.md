---
id: byob-iostreams.2
title: ColorScheme honors NO_COLOR + TTY, degrades to identity function
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

Problem: terminal color codes embedded across the codebase create three
failure modes: (1) garbage in pipes and log files, (2) ignoring the user's
`NO_COLOR` preference, (3) ugly `if isTTY` guards at every print site.

Idea: attach a `ColorScheme` to `IOStreams`. It exposes `Red(s)`, `Green(s)`,
`Bold(s)`, etc. When colors are disabled (non-TTY stdout, `NO_COLOR` env var
set, or explicit `--no-color`), those methods return the input unchanged.
Call sites always write `cs.Red("error")` without guarding.

Tradeoffs: method-per-color grows the surface slightly. Respecting the
`NO_COLOR` standard is table stakes in 2025; costs nothing to honor.

## Design

```go
type ColorScheme struct{ enabled bool }

func NewColorScheme(isTTY bool) *ColorScheme {
    _, noColor := os.LookupEnv("NO_COLOR")
    return &ColorScheme{enabled: isTTY && !noColor}
}

func (c *ColorScheme) Red(s string) string {
    if !c.enabled { return s }
    return "\x1b[31m" + s + "\x1b[0m"
}

// usage:
fmt.Fprintln(io.ErrOut, io.ColorScheme().Red("error: ") + err.Error())
```

