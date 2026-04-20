---
id: byob-89e.2
title: Logger on the Factory, injected into context
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

Problem: two extremes both hurt. Threading `*slog.Logger` as a
parameter through every helper pollutes every signature. Using
`slog.Default()` everywhere makes per-command attributes (command
name, run ID, host) impossible without stepping on other callers'
defaults.

Idea: the Factory holds the root logger as an eager field
(`f.Logger *slog.Logger`) — the logger itself is cheap to construct
(one slog handler over `IO.ErrOut`) so it doesn't need the
`func() (T, error)` lazy-closure shape that byob-1dv.1 reserves for
expensive deps. The root command's
`PersistentPreRunE` (byob-n37.6) attaches per-run attributes with
`logger.With(...)` and stuffs the result into `cmd.Context()` — the
same context already threaded through every runFunc (byob-w71.2).
Command code reaches for the logger via `slog.InfoContext(ctx, ...)`
or a tiny `logs.From(ctx)` helper.

Tradeoffs: `slog.FromContext` is not stdlib — you write a 5-line
context-key helper. Not every log call site has a ctx, so a default
logger still exists as a fallback. The win: commands never pass a
logger parameter, and per-command attributes attach in one place.

## Design

```go
// internal/logs/ctx.go
type ctxKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
    return context.WithValue(ctx, ctxKey{}, l)
}
func From(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}

// pkg/cmd/root/root.go, inside PersistentPreRunE:
l := f.Logger.With("cmd", cmd.CommandPath())
cmd.SetContext(logs.WithLogger(cmd.Context(), l))
```

The `cmd.SetContext` call must run on the leaf command's
`PersistentPreRunE` (cobra sets it on the specific command under
invocation; children inherit at run time). A helper that runs
before cobra resolves the target command sees a stale context and
the logger attribute never lands.

