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
`SilenceUsage` stops the usage dump on every cobra-emitted error;
`SilenceErrors` stops cobra from printing the error string at all — which
you then do yourself in the top-level runner, so you control the
formatting (color, hints, exit code mapping).

The catch: `SilenceUsage = true` cascades to flag-parsing errors too —
"unknown flag", "missing argument", and the unknown-command path no
longer print usage. The runner's "error: ..." line is the only signal
the user sees. Pair this with a `cobra.SetFlagErrorFunc` that wraps
pflag errors as `*FlagError` (so the runner maps them to exit 2) and a
string-prefix check for "unknown command" in the runner (cobra has no
typed sentinel for that path). The exit code carries the "you invoked
me wrong" semantics; the message carries the diagnostic.

If you'd rather have usage on flag errors and silence only on RunE
errors, do the gh-cli inversion: drop `SilenceUsage` from root, and set
`cmd.SilenceUsage = true` as the first line of every `RunE`. Cobra
emits flag errors before `RunE`, so usage still prints there. The
trade-off is a per-command line of boilerplate the runner-on-root
approach avoids.

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
    // Wrap pflag's flag-parse errors so the runner exits 2 per byob-errors.1.
    root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
        return &cmdutil.FlagError{Err: err}
    })
    // ... register groups and subcommands ...
    return root
}

// runner: wrap cobra's untyped "unknown command" message as FlagError
// before mapping. Cobra has no public sentinel for that path.
func classify(err error) error {
    if err == nil { return nil }
    var fe *cmdutil.FlagError
    if errors.As(err, &fe) { return err }
    if strings.HasPrefix(err.Error(), "unknown command ") {
        return &cmdutil.FlagError{Err: err}
    }
    return err
}

// main.go
func main() {
    root := pkgcmd.NewCmdRoot(factory.New())
    err := root.ExecuteContext(ctx)
    os.Exit(runner.MapErrorToExitCode(classify(err)))
}
```

