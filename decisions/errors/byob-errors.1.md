---
id: byob-errors.1
title: Semantic error types; top-level runner maps them to exit codes
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

Problem: commands that call `os.Exit` are untestable and skip deferred
cleanup. Commands that always return `exit 1` give scripts no signal about
what went wrong.

Idea: commands return errors. Define a small vocabulary of error types or
sentinels — `FlagError` (bad flags), `SilentError` (already printed,
suppress), `CancelError` (user cancelled), `NoResultsError` (empty, not
failure). A top-level runner `Run(root, args) int` unwraps these and returns
distinct exit codes. Commands stay oblivious to exit codes.

Tradeoffs: one more tiny package (`cmdutil/errors.go`). You stop writing
`os.Exit(1)` anywhere except in `main()`, which is correct anyway.

## Design

```go
type FlagError struct{ Err error }
func (e *FlagError) Error() string { return e.Err.Error() }
func (e *FlagError) Unwrap() error { return e.Err }

var ErrSilent = errors.New("silent")   // already reported; just exit 1
var ErrCancel = errors.New("cancel")   // user cancelled; exit 2

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

