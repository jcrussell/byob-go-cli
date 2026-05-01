---
id: byob-iu5.2
title: Paths on the Factory as a lazy field
type: decision
priority: 2
status: open
parent: byob-iu5
labels:
- cli
- factory-di
- go
- state
---

## Description

Problem: every command that touches state or cache needs the
resolved paths. Computing them in each command re-runs
`os.UserConfigDir()` etc. redundantly and loses a single test
injection point. A package-level global makes tests non-parallel-safe.

Idea: put `Paths()` on the Factory as a lazy `sync.OnceValue`
closure, mirroring byob-xgz.3 (config), byob-1dv.1 (Factory shape),
and the `sync-oncevalue` memory. First caller pays the resolution
cost; subsequent callers get the cached value. Tests override
`f.Paths` with a struct pointing at `t.TempDir()` subdirectories,
keeping each test hermetic (per the `test-tempdir` memory).

Tradeoffs: one more lazy field on the Factory. The laziness is
load-bearing: it means `mytool --version` and `mytool --help` pay no
filesystem cost at all (aligned with byob-xgz.3's cold-start
argument).

## Design

```go
type Factory struct {
    // ...eager fields (IOStreams, Prompter, Logger)
    Paths func() (*paths.Paths, error)
    // ...other lazy fields
}

func New() *Factory {
    return &Factory{
        // ...
        Paths: sync.OnceValues(func() (*paths.Paths, error) {
            return paths.Resolve("mytool")
        }),
    }
}

// In a command:
p, err := f.Paths()
if err != nil { return err }
if err := paths.EnsureDir(p.Cache); err != nil { return err }
f := filepath.Join(p.Cache, "latest.json")
```

Tests:

```go
func newTestFactory(t *testing.T) *Factory {
    dir := t.TempDir()
    return &Factory{
        Paths: func() (*paths.Paths, error) {
            return &paths.Paths{
                Config: filepath.Join(dir, "config"),
                Cache:  filepath.Join(dir, "cache"),
                State:  filepath.Join(dir, "state"),
            }, nil
        },
        // ...
    }
}
```

