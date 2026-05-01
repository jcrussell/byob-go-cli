---
id: byob-logging.1
title: slog as the logger, stdlib only
type: decision
priority: 2
status: open
parent: byob-logging
labels:
- cli
- go
- logging
---

## Description

Problem: picking a logging library (`logrus`, `zap`, `zerolog`) locks
the whole tool into that library's API, handler model, performance
envelope, and dependency footprint. Migrating later means touching
every log call site. And third-party loggers drift in and out of
maintenance while stdlib does not.

Idea: use `log/slog` from the standard library (Go 1.21+). It ships
with `TextHandler` and `JSONHandler`, has a pluggable `Handler`
interface for custom backends (redaction, routing, testing), supports
`LogAttrs` for a zero-alloc hot path, and composes with
`context.Context` via `slog.InfoContext`. No extra dependency, no
migration risk, and the API is small enough to learn in an afternoon.

Tradeoffs: slog's ecosystem of pre-built handlers (OTLP, Sentry,
file rotation) is thinner than logrus's. You'll write the occasional
custom handler — but the `Handler` interface is four methods, so
that's a weekend project, not a rewrite.

## Design

```go
import "log/slog"

// Root handler, wired to IOStreams.ErrOut (see byob-iostreams.3).
h := slog.NewTextHandler(f.IOStreams.ErrOut, &slog.HandlerOptions{
    Level: slog.LevelWarn, // quiet by default; -v/-vv raises it
})
logger := slog.New(h)
slog.SetDefault(logger)

// In command code:
slog.InfoContext(ctx, "fetching items", "count", len(ids), "host", cfg.Host)
```

