---
id: byob-layout.1
title: 'Three-tier layout: cmd/<bin> -> internal/<bin>cmd -> pkg/cmd/<feature>'
type: decision
priority: 2
status: open
parent: byob-layout
labels:
  - cli
  - go
  - layout
---

## Description

Problem: fat `main.go` files accrue flag parsing, business logic, and error
handling. Single-package `cmd/` trees force circular-import gymnastics when
commands want to share helpers.

Idea: a conventional three-tier layout.
- `cmd/<bin>/main.go` — ~20 lines. Build factory, call runner, exit with
  mapped error code.
- `internal/<bin>cmd/cmd.go` — the runner with error-type → exit-code
  mapping and any process-global concerns (signal handling, profile flag).
- `pkg/cmd/root/root.go` — root cobra command, groups, aggregates features.
- `pkg/cmd/<feature>/<feature>.go` — one NewCmdXxx per feature, each with
  its own `_test.go`.

Adding a new feature is a self-contained PR: new package under `pkg/cmd/`,
one import line in `root.go`.

Tradeoffs: more directories than a flat layout. The flat layout dies around
5 commands; this scales to 50.

## Design

```
cmd/mytool/main.go               // 20 lines
internal/mytoolcmd/cmd.go        // Run(ios) int — error→exit-code mapping
pkg/cmd/root/root.go             // root cobra.Command + groups
pkg/cmdutil/factory.go           // Factory definition
pkg/cmdutil/errors.go            // FlagError, ErrHint, FlagErrorf
pkg/iostreams/iostreams.go       // IOStreams, Test(), ColorScheme
pkg/cmd/create/create.go         // NewCmdCreate(f, runF)
pkg/cmd/create/create_test.go
pkg/cmd/list/list.go
pkg/cmd/list/list_test.go
```

