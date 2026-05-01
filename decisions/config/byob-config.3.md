---
id: byob-config.3
title: Lazy config load behind a factory closure
type: decision
priority: 2
status: open
parent: byob-config
labels:
- cli
- config
- go
---

## Description

Problem: if `main()` always reads, parses, and validates the config file,
`mytool --version` and `mytool --help` pay a filesystem cost and can fail on
a broken config file they don't even need.

Idea: put config loading inside a lazy factory closure. Only commands that
actually dereference `f.Config()` trigger the load. `sync.Once` guarantees
one load per process.

Tradeoffs: you must remember `f.Config()` (invoke) instead of a bare field.
Alternative: eager load. Simpler to reason about, slower to start, and
turns "broken config file" into "can't see --help".

## Design

```go
func (f *Factory) configProvider() func() (*Config, error) {
    var (
        once sync.Once
        cfg  *Config
        err  error
    )
    return func() (*Config, error) {
        once.Do(func() {
            path, e := FindConfigUp("mytool.toml")
            if e != nil { cfg, err = defaultConfig(), nil; return }
            cfg, err = loadAndValidate(path)
        })
        return cfg, err
    }
}

// main() never reads config; 'mytool --help' is filesystem-free.
```

