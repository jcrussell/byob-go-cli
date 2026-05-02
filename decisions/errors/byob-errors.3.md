---
id: byob-errors.3
title: FlagErrorf helper for value-validation errors that map to exit code 2
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

Problem: runtime validation of flag *values* ("--port must be 1-65535",
"--format must be one of json|yaml|table") needs to return an error that
the top-level runner maps to exit code 2 (the "you invoked me wrong" exit
code), not the default exit code 1 (generic failure). If runFunc returns
a plain `fmt.Errorf("...")`, the runner has no way to distinguish it from
a runtime error and will emit exit 1.

Idea: a small `cmdutil.FlagErrorf(format, args...)` helper that wraps
`fmt.Errorf` with a `*FlagError` type. The top-level runner matches on
`*FlagError` via `errors.As` and maps it to exit code 2. runFuncs use
`cmdutil.FlagErrorf(...)` for any validation that produces a "bad flag
value" error, regardless of whether cobra's built-in flag-group helpers
could have caught it earlier.

For flag *relationships* — mutually exclusive, required together,
one-of-N required — use cobra's declarative helpers instead:
`cmd.MarkFlagsMutuallyExclusive("json", "yaml", "template")`,
`cmd.MarkFlagsRequiredTogether("key", "secret")`,
`cmd.MarkFlagsOneRequired("file", "stdin", "url")`. They run in cobra's
validation phase before runFunc, emit consistent error messages, and
integrate with shell completion. See the `cobra-flag-groups` memory for
the quick reminder.

Tradeoffs: `FlagErrorf` is a thin wrapper. Its whole value is in the
*type* being `*FlagError`, which the runner uses. Without the runner's
type assertion, `FlagErrorf` adds nothing; with it, you get consistent
exit-code-2 behavior across every command without each runFunc knowing
anything about exit codes.

When not to use: if you're not on cobra and the top-level runner doesn't
map error types to exit codes, skip the helper and use plain
`fmt.Errorf`.

## Design

```go
// cmdutil/errors.go
type FlagError struct{ Err error }
func (e *FlagError) Error() string { return e.Err.Error() }
func (e *FlagError) Unwrap() error { return e.Err }

func FlagErrorf(format string, args ...any) error {
    return &FlagError{Err: fmt.Errorf(format, args...)}
}

// Inside a runFunc:
if opts.Port < 1 || opts.Port > 65535 {
    return cmdutil.FlagErrorf("--port must be 1-65535, got %d", opts.Port)
}
if _, ok := validFormats[opts.Format]; !ok {
    return cmdutil.FlagErrorf(
        "--format must be one of json|yaml|table, got %q", opts.Format,
    )
}

// Top-level runner (in internal/<bin>cmd/cmd.go):
func Run(root *cobra.Command, args []string) int {
    err := root.Execute()
    var flagErr *cmdutil.FlagError
    switch {
    case err == nil:
        return 0
    case errors.As(err, &flagErr):
        fmt.Fprintln(os.Stderr, "error:", err)
        return 2
    default:
        fmt.Fprintln(os.Stderr, "error:", err)
        return 1
    }
}
```

