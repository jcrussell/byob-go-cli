---
id: byob-config.1
title: Walk-up-from-cwd to find project-scoped config
type: decision
priority: 2
status: open
parent: byob-config
labels:
  - config
---

## Description

Problem: project-scoped config files need to work from any subdirectory, not
just the project root. Users expect to run `mytool` anywhere inside their
project and have it find the right config.

Idea: starting from `os.Getwd()`, walk upward toward the filesystem root
looking for a config file (`mytool.toml`, `.mytoolrc`, etc). The first match
wins. If none is found before hitting root, fall back to `~/.config/mytool/`
or a pure-defaults config.

Tradeoffs: one filesystem stat per ancestor directory. Negligible for depths
seen in practice. Same UX users already expect from git, cargo, and direnv.
Avoid: searching XDG dirs *before* the walk — project config must always
win.

## Design

```go
func FindConfigUp(name string) (string, error) {
    dir, err := os.Getwd()
    if err != nil { return "", err }
    for {
        p := filepath.Join(dir, name)
        if _, err := os.Stat(p); err == nil {
            return p, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir { // reached filesystem root
            return "", fs.ErrNotExist
        }
        dir = parent
    }
}
```

