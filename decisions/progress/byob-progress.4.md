---
id: byob-progress.4
title: Progress writes to ErrOut, never Out
type: decision
priority: 2
status: open
parent: byob-progress
labels:
  - iostreams
  - progress
---

## Description

Problem: a spinner or bar written to stdout destroys pipeability.
`mytool fetch | jq .` gets `\r`-animated ANSI codes in its input
stream, and the `jq` parse fails — or worse, silently succeeds on
garbage.

Idea: progress writes **only** to `IO.ErrOut`. This reinforces
byob-iostreams.3 (data to Out, chatter to ErrOut) and byob-output.1
(TTY-adaptive output goes to the adaptive side). Data — the thing
a user might pipe — stays on Out, uncluttered. Chatter + progress
share ErrOut; the byob-logging.4 "quiet by default" posture keeps logs
out of the way unless asked.

Rule of thumb: if removing a print would change what a pipe
consumer sees, that print belongs to Out. Progress indicators never
qualify — their entire job is ephemeral status.

Tradeoffs: none worth noting. This is a corollary of byob-iostreams.3
more than a new decision.

## Design

```go
// Factory method always wires to ErrOut:
func (f *Factory) Progress(ctx context.Context, label string) progress.Progress {
    if f.IOStreams.IsStderrTTY() {
        return progress.NewSpinner(ctx, f.IOStreams.ErrOut, label) // NOT Out
    }
    return progress.NewLogging(ctx, f.IOStreams.ErrOut, label)
}

// Consumer code stays pipe-safe:
p := f.Progress(ctx, "fetching")
p.Start()
items := fetch()
p.Stop()
for _, item := range items {
    fmt.Fprintln(f.IOStreams.Out, item.Name) // data to Out
}
```

