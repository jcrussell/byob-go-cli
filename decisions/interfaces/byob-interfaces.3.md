---
id: byob-interfaces.3
title: Accept fs.FS at package boundaries for filesystem seams
type: decision
priority: 2
status: open
parent: byob-interfaces
labels:
  - cli
  - go
  - interfaces
---

## Description

Problem: a package that reads files via `os.Open` or `os.ReadFile` has
baked disk I/O into its API. Tests have to create temp directories,
write fixture files, and clean up — or reach for path-rewriting hacks.
Embedded assets end up on a parallel code path.

Idea: accept `fs.FS` at package boundaries instead of string paths.
`fs.FS` is a tiny stdlib interface (`Open(name string) (fs.File,
error)`), and `fs.ReadFile`, `fs.WalkDir`, `fs.Glob`, and `fs.Sub` all
take one. In production, pass `os.DirFS("/path")`. In tests, pass
`fstest.MapFS{}` — an in-memory fake with no disk touched, no temp
dirs, no cleanup. For embedded assets, `embed.FS` already satisfies
`fs.FS`, so the same code path works for on-disk files, embedded
resources, and hermetic tests.

Tradeoffs: the interface is read-only. If your package also writes
files, keep the write path separate (accept a writer or a directory
path for writes). That's a clean split, not a leak — read and write
really are different concerns.

When not to use: for purely invocation-level file reads at a runFunc
boundary (e.g., `cmd foo --config /path/to/file`), parsing the path
and handing the parsed value down is fine. The seam is for packages
that do non-trivial filesystem work internally.

## Design

```go
package assets

import "io/fs"

type Loader struct {
    FS fs.FS
}

func (l *Loader) Load(name string) ([]byte, error) {
    return fs.ReadFile(l.FS, name)
}

// production wiring
//   loader := &assets.Loader{FS: os.DirFS("/etc/myapp")}
//
// embedded wiring
//   //go:embed defaults/*
//   var defaults embed.FS
//   loader := &assets.Loader{FS: defaults}
//
// test wiring
//   loader := &assets.Loader{FS: fstest.MapFS{
//       "config.toml": &fstest.MapFile{Data: []byte(`key = "value"`)},
//   }}
```

