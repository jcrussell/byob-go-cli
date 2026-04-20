---
id: byob-89e.3
title: 'Verbosity ladder: -v / -vv / --log-level on root'
type: decision
priority: 2
status: open
parent: byob-89e
labels:
- cli
- go
- logging
---

## Description

Problem: users want a quick "more output" knob (`-v`), a "full debug"
knob (`-vv`), AND explicit control for scripts (`--log-level=info`).
Picking just one frustrates the others. Env-var control
(`MYTOOL_LOG=debug`) is expected by anyone who's used `gh` or
`kubectl`.

Idea: three persistent flags on root (byob-n37.6 pattern) with a
documented precedence:

1. `--log-level=warn|info|debug` — explicit, wins when set.
2. `-v`/`-vv` — count flag: `-v` → Info, `-vv`+ → Debug. Loses to
   `--log-level` when both are passed.
3. `MYTOOL_LOG=<level>` env var — loses to both flags, beats the
   default.
4. Default — `Warn`. The binary is quiet unless asked otherwise.

The "flags beat env" direction is deliberately opposite to byob-xgz.2
(which says env > file > default for *config*). Logging verbosity is
per-invocation, not per-shell; a `-vv` on the command line should not
be silenced by a stale `MYTOOL_LOG=warn` in the environment.

Tradeoffs: three knobs is one more than the lean option. Users who
only want `-v`/`-vv` can ignore the other two. Scripting users want
the explicit `--log-level` because `-vv` reads like a typo in a CI
log.

## Design

```go
// pkg/cmd/root/root.go
var verbose int
var logLevel string
root.PersistentFlags().CountVarP(&verbose, "verbose", "v",
    "increase log verbosity (-v=info, -vv=debug)")
root.PersistentFlags().StringVar(&logLevel, "log-level", "",
    "explicit log level (warn|info|debug); overrides -v")

// In PersistentPreRunE.
// Note: cobra's CountVarP starts at 0 and has no way to express
// "explicit -v0", so verbose == 0 and "no -v passed" are the same
// state. That means env-var MYTOOL_LOG wins when no -v/-vv flag is
// present — which is the intended precedence.
lvl := slog.LevelWarn
switch {
case logLevel != "":
    lvl = parseLevel(logLevel) // warn|info|debug
case verbose >= 2:
    lvl = slog.LevelDebug
case verbose == 1:
    lvl = slog.LevelInfo
case os.Getenv("MYTOOL_LOG") != "":
    lvl = parseLevel(os.Getenv("MYTOOL_LOG"))
}
```

