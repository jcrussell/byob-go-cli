---
id: byob-lifecycle.3
title: '`errgroup` as the goroutine-fanout default'
type: byob
priority: 2
status: open
parent: byob-lifecycle
labels:
  - concurrency
  - context
  - lifecycle
---

## Description

Problem: a CLI command that fans out N parallel calls (concurrent HTTP
fetches, parallel store reads, multiple subprocess invocations) needs
ctx cancellation, first-error short-circuit, and bounded concurrency.
Hand-rolling that with `sync.WaitGroup` + an error channel +
`context.WithCancel` invariably gets the cancellation ordering wrong on
first attempt and re-derives the same primitive across every command.

Idea: byob default for "fan out N tasks, wait, surface the first error"
is `golang.org/x/sync/errgroup` with a context-aware Group. Pick up the
ctx threaded by byob-lifecycle.1, derive a Group with
`errgroup.WithContext(ctx)`, spawn with `g.Go(func() error { ... })`,
collect with `g.Wait()`. The first non-nil error cancels the derived
ctx; every other goroutine sees `ctx.Done()` and returns. Bounded
concurrency via `g.SetLimit(n)`.

Pairs with the `goroutine-exit-path` memory: errgroup is the byob
implementation of "every goroutine has a documented exit path" for the
fanout case.

Tradeoffs: stdlib `sync.WaitGroup` + error channel works for one-off
cases and avoids a (small, well-maintained) third-party dep on
`golang.org/x/sync`. The dep is `golang.org/x` adjacent-stdlib, used
by Kubernetes, gh CLI, and most production Go — not a foreign dep. The
packaging cost is one go.mod line; the wiring cost saved is several
lines per fanout site, plus the cancellation-ordering bugs you don't
write.

When not to use:
- Long-lived background goroutines (process-lifetime workers): use
  plain ctx + select + a parent shutdown channel.
- A single goroutine spawned for parallelism inside one blocking call:
  a plain `go` + result channel is fine.
- "Fire N tasks, wait for all, collect every error" — errgroup
  short-circuits on first error. Use a `[]error` slice plus
  `errors.Join` if you need all of them.

## Design

```go
import (
    "context"
    "fmt"

    "golang.org/x/sync/errgroup"
)

func parallelFetch(ctx context.Context, urls []string) ([]Result, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(8)  // bounded concurrency

    out := make([]Result, len(urls))
    for i, u := range urls {
        g.Go(func() error {
            r, err := fetch(ctx, u)  // ctx cancels on first error
            if err != nil {
                return fmt.Errorf("fetch %q: %w", u, err)
            }
            out[i] = r
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return out, nil
}
```

ctx threading per byob-lifecycle.1 supplies the parent context; errgroup's
derived ctx cancels on first error so in-flight goroutines abort rather
than running to completion only to have their results discarded. Loop
variables don't need shadowing in Go 1.22+.

