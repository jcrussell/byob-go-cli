---
id: byob-w71.2
title: Thread context.Context through every runFunc
type: decision
priority: 2
status: open
parent: byob-w71
labels:
- cli
- go
- lifecycle
---

## Description

Problem: a runFunc signature of `func(opts *Options) error` can't be
cancelled. Every HTTP call, DB query, or subprocess inside it is
uninterruptible, and Ctrl-C produces an orphaned child process or a stuck
database connection.

Idea: follow the Go stdlib convention: every function that might block
takes a `context.Context` as its first argument, and returns quickly when
`ctx.Done()` fires. runFuncs are no exception — make them
`func(ctx context.Context, opts *Options) error`.

`main()` constructs the root ctx (see byob-w71.3 for the
`signal.NotifyContext` setup) and threads it through
`root.ExecuteContext(ctx)`. Cobra's `cmd.Context()` returns that same
context inside every RunE, so the plumbing is: pull `ctx` out of
`cmd.Context()` and pass it to your runFunc.

Tradeoffs: every blocking call inside a command now has to accept ctx and
pass it through. That's a one-time audit of your codebase, not ongoing
friction. The alternative — ignoring cancellation — produces the
user-hostile behavior described in the Problem.

When not to use: never. Context threading is table stakes in modern Go.

## Design

```go
// main.go — see byob-w71.3 for the signal.NotifyContext wiring that
// feeds root.ExecuteContext(ctx). This decision picks up from there.

// pkg/cmd/list/list.go
type Options struct {
    IO    *iostreams.IOStreams
    Store func() (Store, error)
    Limit int
}

func NewCmdList(f *Factory, runF func(ctx context.Context, opts *Options) error) *cobra.Command {
    opts := &Options{IO: f.IOStreams, Store: f.Store}
    cmd := &cobra.Command{
        Use: "list",
        RunE: func(c *cobra.Command, args []string) error {
            if runF != nil { return runF(c.Context(), opts) }
            return listRun(c.Context(), opts)
        },
    }
    return cmd
}

func listRun(ctx context.Context, opts *Options) error {
    s, err := opts.Store()
    if err != nil { return err }
    items, err := s.List(ctx) // cancellable
    ...
}
```

