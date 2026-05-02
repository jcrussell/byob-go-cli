---
id: byob-dependencies.1
title: 'Pure-Go discipline: CGO_ENABLED=0, pure-Go drivers, go:embed assets'
type: decision
priority: 2
status: open
parent: byob-dependencies
labels:
  - cli
  - deps-philosophy
  - go
---

## Description

Problem: CGO makes cross-compilation a toolchain puzzle. Non-embedded assets
mean shipping a tarball instead of a binary. Users hit "install
libsqlite3-dev" on day one.

Idea: build with `CGO_ENABLED=0`. Use pure-Go implementations for anything
that traditionally required C — `modernc.org/sqlite` for SQLite, pure-Go
crypto, pure-Go tls. Ship every asset (SQL migrations, templates, static
files) via `go:embed`.

Payoff: one binary, trivial cross-compile (`GOOS=darwin GOARCH=arm64 go
build`), no install docs beyond "download and run", no runtime filesystem
layout to document.

Tradeoffs: some pure-Go drivers are slower than their C counterparts. For
typical CLI workloads — opening a sqlite file, running a few hundred
queries — the gap is invisible. Benchmark if you have a hot loop.

## Design

```go
// go.mod
require (
    modernc.org/sqlite v1.x.x
    github.com/spf13/cobra v1.x.x
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed templates/*.tmpl
var templatesFS embed.FS

// Makefile:
//   CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" ./cmd/mytool
//   GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build ...
//   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ...
```

