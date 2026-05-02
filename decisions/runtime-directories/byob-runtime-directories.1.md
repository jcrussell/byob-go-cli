---
id: byob-runtime-directories.1
title: 'Four-directory model: config / cache / state / data'
type: decision
priority: 2
status: open
parent: byob-runtime-directories
labels:
  - cli
  - go
  - state
---

## Description

Problem: CLIs that dump everything into `~/.mytool/` conflate
user-editable config with regenerable cache, persistent state, and
shipped data. Cleaning a user's cache then destroys their config.
Following XDG on Linux but not macOS/Windows means platform users
get inconsistent experiences.

Idea: three distinct directories, each resolved per-OS via the
stdlib where possible and an in-repo helper for state (the stdlib
has no `UserStateDir`):

- **config** — `os.UserConfigDir()` (covered by byob-config for the
  walk-up discovery). User-editable, small, rarely changes.
  Linux: `~/.config/mytool/`; macOS: `~/Library/Application Support/mytool/`;
  Windows: `%AppData%\mytool\`.
- **cache** — `os.UserCacheDir()` joined with tool name. Regenerable,
  safe to `rm -rf`. Fetched data, compiled templates, thumbnail
  blobs, anything the tool can recreate on demand. Windows:
  `%LocalAppData%\mytool\`.
- **state** — an in-repo `paths.stateDir()` helper joined with tool
  name. Non-user-editable but persistent: last-used-repo, undo
  history, auth state (when added later), run counters. Lost-state
  is annoying but not destructive.
  - Unix: `$XDG_STATE_HOME` → `$HOME/.local/state`.
  - macOS: `$HOME/Library/Application Support/<tool>` (shared with
    config; distinguish by subfolder).
  - Windows: `%LocalAppData%\<tool>` (same directory as cache;
    distinguish by subfolder).

Application-shipped templates or plugins ("data") aren't a separate
category here — when a tool needs them, pick cache or state per
regenerability and nest under the chosen root.

Each directory is `MkdirAll`'d on first use, not at startup — a
plain `mytool --version` touches zero directories.

Tradeoffs: three path concepts instead of one. Worth it: a
well-shaped tool can answer "where does X live?" from category
alone, and users can `rm -rf $(mytool paths cache)` confident they
haven't lost config. The stdlib lacks a `UserStateDir`, so the
state resolver is 30 lines of `runtime.GOOS` branching — the one
wart of this scheme.

## Design

```go
// pkg/cmd/paths/paths.go
type Paths struct {
    Config string
    Cache  string
    State  string
}

func Resolve(toolName string) (*Paths, error) {
    cfg, err := os.UserConfigDir()
    if err != nil { return nil, err }
    cache, err := os.UserCacheDir()
    if err != nil { return nil, err }
    state, err := stateDir()
    if err != nil { return nil, err }
    p := &Paths{
        Config: filepath.Join(cfg, toolName),
        Cache:  filepath.Join(cache, toolName),
        State:  filepath.Join(state, toolName),
    }
    // stateDir() returns a shared root with Config on macOS and with
    // Cache on Windows. When the joined paths collide, nest State
    // under a dedicated subdir so the three categories are always
    // distinct.
    if p.State == p.Config || p.State == p.Cache {
        p.State = filepath.Join(p.State, "state")
    }
    return p, nil
}

// stateDir returns the per-OS root for persistent, non-regenerable
// state. The stdlib has UserConfigDir / UserCacheDir but no
// UserStateDir, so this is hand-rolled.
func stateDir() (string, error) {
    if runtime.GOOS == "windows" {
        if d := os.Getenv("LocalAppData"); d != "" {
            return d, nil
        }
        return os.UserCacheDir() // same root as cache on Windows
    }
    if runtime.GOOS == "darwin" {
        home, err := os.UserHomeDir()
        if err != nil { return "", err }
        return filepath.Join(home, "Library", "Application Support"), nil
    }
    // Unix: XDG_STATE_HOME, else ~/.local/state.
    if d := os.Getenv("XDG_STATE_HOME"); d != "" {
        return d, nil
    }
    home, err := os.UserHomeDir()
    if err != nil { return "", err }
    return filepath.Join(home, ".local", "state"), nil
}

// First-use mkdir lives in the helper that writes to the dir,
// not in Resolve(). Keeps --version free of filesystem writes.
// Callers that are about to write call EnsureDir(p) first — e.g.
// a command that persists auth state runs
// `paths.EnsureDir(p.State)` then `paths.WriteFileAtomic(...)`
// (byob-runtime-directories.3).
func EnsureDir(p string) error { return os.MkdirAll(p, 0o755) }
```

