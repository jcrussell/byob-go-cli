---
id: byob-progress.1
title: Progress as a narrow interface on the Factory
type: decision
priority: 2
status: open
parent: byob-progress
labels:
  - factory-di
  - interfaces
  - progress
---

## Description

Problem: every spinner/bar library (`charmbracelet/bubbles`,
`schollz/progressbar`, `briandowns/spinner`) exposes a different
shape. Baking one into command code couples every caller to that
library and makes the test path awkward.

Idea: a narrow interface constructed per-operation via the Factory
(byob-interfaces.1). Not a field — a factory method, because each progress
instance is operation-scoped. The interface has four methods:
`Start()`, `Update(msg string)`, `Stop()`, `Fail(err error)`.
`Stop` and `Fail(err)` are mutually-exclusive terminal states —
exactly one is called per operation, after any number of `Update`s.
Calling either after the operation has already terminated is a no-op.
Consumer code is library-agnostic; the interface lives in
`pkg/cmd/progress/`.

Factory shape: `f.Progress(ctx context.Context, label string)
progress.Progress`. The returned instance encapsulates the
TTY-vs-logging decision (byob-progress.2) internally so callers don't
branch on it. The `ctx` is captured at construction and watched
inside the impl: if it cancels mid-operation the impl calls its
own `Stop()` so the spinner doesn't keep rendering under a dead
command. Happy-path `Stop`/`Fail` still belongs to the caller
(usually `defer p.Stop()` right after `p.Start()`).

Tradeoffs: four methods is deliberately thin. No nested sub-progress,
no multi-line bars, no simultaneous concurrent progresses. Tools
that need those add them as separate interfaces (`ProgressGroup`),
not by stretching this one.

## Design

```go
// pkg/cmd/progress/progress.go
package progress

type Progress interface {
    Start()
    Update(msg string)
    Stop()
    Fail(err error)
}

// pkg/cmdutil/factory.go
func (f *Factory) Progress(ctx context.Context, label string) progress.Progress {
    if f.IOStreams.IsStderrTTY() {
        return progress.NewSpinner(ctx, f.IOStreams.ErrOut, label)
    }
    return progress.NewLogging(ctx, f.IOStreams.ErrOut, label)
}

// Usage:
p := f.Progress(ctx, "fetching items")
p.Start()
defer p.Stop()
for i, id := range ids {
    p.Update(fmt.Sprintf("item %d/%d", i+1, len(ids)))
    // ...
}
```

For known-total progress bars, a sibling method:
`f.ProgressBar(ctx context.Context, label string, total int)
progress.Progress`. Same interface, different constructor,
different library underneath.

