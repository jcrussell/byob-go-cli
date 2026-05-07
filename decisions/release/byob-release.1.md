---
id: byob-release.1
title: Version via -ldflags -X into a dedicated build package
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - concurrency
  - release
---

## Description

Problem: the binary needs to know its own version, commit, and build
date at runtime. Options: ldflags-injected package vars,
`debug.ReadBuildInfo()`, a hand-maintained `version.go` constant, a
generated file. Each has gaps — constants go stale, generated files
churn, BuildInfo is empty or partial under `go build` from a dirty
tree or from an archive extraction.

Idea: a small package `internal/<bin>cmd/build` exposes three
package-scope vars with sentinel defaults (named constants so
consumers test against one symbol, not a string literal):

```go
const (
    VersionDev  = "dev"
    CommitNone  = "none"
    DateUnknown = "unknown"
)
var (
    Version = VersionDev
    Commit  = CommitNone
    Date    = DateUnknown
)
```

Both the Makefile and `.goreleaser.yml` inject the same ldflags so
every build path populates the same three vars:

```
-ldflags "-X <pkg>/build.Version=$(VERSION)
         -X <pkg>/build.Commit=$(COMMIT)
         -X <pkg>/build.Date=$(DATE)"
```

`debug.ReadBuildInfo()` is a **secondary** source inside the build
package: when `Version == VersionDev` (no ldflags), an `Info()`
accessor consults BuildInfo for VCS revision + dirty flag so plain
`go install github.com/x/y@latest` still produces a usable `version`
output.

Thread-safety: `Info()` must not mutate the package vars from its
read path. Two parallel test packages calling `build.Info()`
concurrently would race. Compute the augmented values into locals
and return them; wrap in `sync.OnceValue` (single-return — the
closure never errors) so the `debug.ReadBuildInfo` scan runs at
most once per process.

Tradeoffs: ldflags strings are verbose and easy to typo — hence the
shared Makefile variable both the local `build` target and the GHA
release workflow reference. A single source of truth for the flag
string prevents drift between dev and release builds.

## Design

```go
// internal/mytoolcmd/build/build.go
package build

import (
    "runtime/debug"
    "sync"
)

const (
    VersionDev  = "dev"
    CommitNone  = "none"
    DateUnknown = "unknown"
)

var (
    Version = VersionDev
    Commit  = CommitNone
    Date    = DateUnknown
)

type BuildInfo struct {
    Version, Commit, Date string
}

// Info returns the version/commit/date resolved from either ldflags
// (authoritative) or debug.ReadBuildInfo() (fallback for go install /
// go run). Safe for concurrent calls. sync.OnceValue (single-return —
// the closure never errors) so the BuildInfo scan runs at most once.
var Info = sync.OnceValue(func() BuildInfo {
    v, c, d := Version, Commit, Date
    if v == VersionDev {
        if info, ok := debug.ReadBuildInfo(); ok {
            for _, s := range info.Settings {
                switch s.Key {
                case "vcs.revision":
                    if c == CommitNone { c = s.Value }
                case "vcs.time":
                    if d == DateUnknown { d = s.Value }
                case "vcs.modified":
                    if s.Value == "true" { c += "-dirty" }
                }
            }
        }
    }
    return BuildInfo{Version: v, Commit: c, Date: d}
})
```

Callers read `info := build.Info()` and use `info.Version`,
`info.Commit`, `info.Date` directly — no destructure helper needed.

Makefile exports `LDFLAGS` once and both `build` and `release`
targets consume it; `.goreleaser.yml` uses `{{ .Env.LDFLAGS }}` or
its own template to produce the same `-X` entries. No init() in this
package — keep `go test -run -short` hermetic and avoid ordering
surprises.

