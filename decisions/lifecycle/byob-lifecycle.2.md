---
id: byob-lifecycle.2
title: Wire Ctrl-C / SIGTERM via signal.NotifyContext in main
type: decision
priority: 2
status: open
parent: byob-lifecycle
labels:
  - cli
  - go
  - lifecycle
---

## Description

Problem: users hit Ctrl-C and expect the CLI to stop. Naive
implementations either ignore the signal (process gets killed
uncleanly, orphaning subprocesses and DB connections) or reach for
channels, goroutines, and shutdown flags — all of which have to be
plumbed into every blocking call.

Idea: use `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`
at the top of `main()`. It returns a `context.Context` that cancels
when either signal fires. Thread that context into
`root.ExecuteContext(ctx)`, and every downstream HTTP call, DB query,
or subprocess that accepts a `context.Context` cancels automatically
— no channels, no goroutines, no shutdown flags.

Defer `cancel()` immediately after to release the signal handler on
clean exit. A second Ctrl-C then reverts to the default handler and
force-exits, which is the right behavior: the first one asks politely;
the second kills.

Tradeoffs: the catch is that your runFuncs and everything they call
must actually thread `ctx` through (see byob-lifecycle.1). If they don't,
the context cancels but nobody listens. That's a one-time audit.

When not to use: never for a production CLI. Every CLI should exit
cleanly on Ctrl-C.

## Design

```go
func main() {
    ctx, cancel := signal.NotifyContext(
        context.Background(), os.Interrupt, syscall.SIGTERM,
    )
    defer cancel()

    root := pkgcmd.NewCmdRoot(factory.New())
    if err := root.ExecuteContext(ctx); err != nil {
        os.Exit(exitCodeFor(err))
    }
}
```

