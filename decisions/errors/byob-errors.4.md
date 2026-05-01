---
id: byob-errors.4
title: SilenceUsage and SilenceErrors on the root command
type: decision
priority: 2
status: open
parent: byob-errors
labels:
- cli
- errors
- go
---

## Description

Problem: the single most common cobra papercut. You return a runtime error
from `RunE` — "connection refused", "file not found", "quota exceeded" —
and cobra prints the entire usage blob (description, flags, subcommands)
followed by the error message. Users hate this: the signal is drowned in
boilerplate, and the usage blob is pointless because the problem isn't
that they misinvoked the command.

Idea: set `SilenceUsage = true` and `SilenceErrors = true` on the root
command. Both settings cascade to every child via inheritance.
`SilenceUsage` stops the usage dump on `RunE` errors. `SilenceErrors`
stops cobra from printing the error string at all — which you then do
yourself in the top-level runner, so you control the formatting (color,
hints, exit code mapping).

Cobra still prints usage on actual flag-parsing errors (unknown flags,
missing arguments), which is the correct behavior — those errors really
are "you invoked me wrong".

Tradeoffs: `SilenceErrors = true` means you're responsible for printing
errors yourself. That's what your top-level runner already does anyway if
you're following the semantic-error-types + ErrHint patterns — those
patterns assume you own the output.

When not to use: never. If you're following the rest of this library, the
combination is strictly better than cobra's defaults.

## Design

```go
func NewCmdRoot(f *Factory) *cobra.Command {
    root := &cobra.Command{
        Use:   "mytool",
        Short: "do the thing",

        // Don't dump usage on runtime errors; users already know how to invoke it.
        SilenceUsage:  true,
        // Don't print errors; the top-level runner formats them.
        SilenceErrors: true,
    }
    // ... register groups and subcommands ...
    return root
}

// main.go
func main() {
    root := pkgcmd.NewCmdRoot(factory.New())
    err := root.ExecuteContext(ctx)
    os.Exit(runner.MapErrorToExitCode(err))  // owns all stderr formatting
}
```

