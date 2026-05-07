---
id: byob-errors.1
title: Semantic error types and a FlagErrorf helper; top-level runner maps them to exit codes
type: decision
priority: 2
status: open
parent: byob-errors
labels:
  - errors
---

## Description

Problem: commands that call `os.Exit` are untestable and skip deferred
cleanup. Commands that always return `exit 1` give scripts no signal about
what went wrong.

Idea: commands return errors. Define a small vocabulary of error types or
sentinels — `FlagError` (bad flags), `SilentError` (already printed,
suppress), `CancelError` (user cancelled), `NoResultsError` (empty, not
failure). A top-level runner `Run(root, args) int` unwraps these and returns
distinct exit codes. Commands stay oblivious to exit codes.

For runtime validation of flag *values* — e.g., `--port` must be 1-65535,
`--format` must be one of `json|yaml|table` — pair `*FlagError` with a
`cmdutil.FlagErrorf(format, args...)` helper that wraps `fmt.Errorf` in
the type. The runner already maps `*FlagError` to exit code 2; without
the wrapper, a plain `fmt.Errorf` is indistinguishable from a runtime
error and lands as exit code 1. (For flag *relationships* — mutually
exclusive, required-together, one-of-N-required — use cobra's declarative
helpers instead. See `byob-command-shape.6`.)

Tradeoffs: one more tiny package (`cmdutil/errors.go`). You stop writing
`os.Exit(1)` anywhere except in `main()`, which is correct anyway.
`FlagErrorf` is a thin wrapper whose whole value is the *type* it
returns; if you're not on cobra and don't have a type-aware top-level
runner, skip the helper and use plain `fmt.Errorf`.

## Design

```go
// cmdutil/errors.go
type FlagError struct{ Err error }
func (e *FlagError) Error() string { return e.Err.Error() }
func (e *FlagError) Unwrap() error { return e.Err }

func FlagErrorf(format string, args ...any) error {
    return &FlagError{Err: fmt.Errorf(format, args...)}
}

var ErrSilent = errors.New("silent")   // already reported; just exit 1
var ErrCancel = errors.New("cancel")   // user cancelled; exit 2

// Inside a runFunc:
//   if opts.Port < 1 || opts.Port > 65535 {
//       return cmdutil.FlagErrorf("--port must be 1-65535, got %d", opts.Port)
//   }

func Run(root *cobra.Command, args []string) int {
    root.SetArgs(args)
    err := root.Execute()
    switch {
    case err == nil:
        return 0
    case errors.Is(err, ErrCancel):
        return 2
    case errors.Is(err, ErrSilent):
        return 1
    case errors.As(err, new(*FlagError)):
        fmt.Fprintln(os.Stderr, err)
        return 2
    default:
        fmt.Fprintln(os.Stderr, "error:", err)
        return 1
    }
}
```

