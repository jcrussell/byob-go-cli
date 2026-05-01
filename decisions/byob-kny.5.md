---
id: byob-kny.5
title: Function signatures return `error`, not the concrete error type
type: decision
priority: 2
status: open
parent: byob-kny
labels:
- cli
- errors
- go
---

## Description

Problem: a constructor like `func newFlag() *FlagError` looks reasonable
— caller gets the typed value, can read fields directly, no `errors.As`
round-trip needed. The trap is that callers who assign the result to
`var err error` get a non-nil interface wrapping a typed-nil pointer
when the constructor returns `nil`. The classic Go gotcha:

```go
func newFlag(condition bool) *FlagError {
    if !condition {
        return nil  // typed nil pointer
    }
    return &FlagError{...}
}

var err error = newFlag(false)
if err != nil { ... }  // ALWAYS true — interface holds (*FlagError)(nil)
```

Idea: functions that can fail return `error`, not the concrete type.
Concrete error *types* are still correct and required — byob-kny.1
explicitly defines `FlagError`, `SilentError`, `CancelError`, and the
top-level runner uses `errors.As` to unwrap them. The rule applies to
the *signature*: return the interface, instantiate the concrete type
only for the value being constructed.

Tradeoffs: callers who need type-specific fields call
`errors.As(err, new(*FlagError))`. That's a one-line round-trip — it's
the mechanism `errors.As` exists for, and the runner in byob-kny.1
already uses it. The Code Review Comments wiki ("avoid in-band error
values") and the Google Go Style decisions ("export error types, return
error") both name this.

When not to use: never. The typed-nil-in-interface trap has zero upside
and the failure mode is a runtime nil-pointer dereference hours after
the apparent error check.

## Design

```go
// Wrong: signature returns *FlagError.
func validate(flags Flags) *FlagError {
    if flags.Name == "" {
        return &FlagError{Err: errors.New("name required")}
    }
    return nil  // typed nil — non-nil when assigned to error
}

// Right: signature returns error; value is *FlagError when non-nil.
func validate(flags Flags) error {
    if flags.Name == "" {
        return &FlagError{Err: errors.New("name required")}
    }
    return nil  // untyped nil — nil interface
}

// Caller unwraps when they need type-specific fields.
if err := validate(opts.Flags); err != nil {
    var ferr *FlagError
    if errors.As(err, &ferr) {
        return ferr  // exit code 2 path in byob-kny.1
    }
    return err
}
```

Same rule applies to any function whose only failure mode is one
sentinel: `func newSilent() error` not `func newSilent() *SilentError`.

