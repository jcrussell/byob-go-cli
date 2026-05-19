---
id: byob-errors.2
title: Attachable ErrHint wrapper carries remediation text
type: byob
priority: 2
status: open
parent: byob-errors
labels:
  - errors
---

## Description

Problem: raw error strings leave users stuck. "permission denied" is accurate
and useless. They file an issue, or they abandon the tool.

Idea: wrap errors at their origin with a short actionable hint. `ErrHint{Err,
Hint}` carries a remediation message ("try `mytool auth login`", "set FOO_DIR
in your environment", "run `mytool doctor` to diagnose"). The top-level
runner prints the underlying error and, on a new line, prints the hint.

Tradeoffs: authors need to remember to attach hints at failure points — most
common ones are worth the discipline. Avoid hint spam; attach only when the
remediation is specific.

## Design

```go
type ErrHint struct {
    Err  error
    Hint string
}
func (e *ErrHint) Error() string { return e.Err.Error() }
func (e *ErrHint) Unwrap() error { return e.Err }

// at failure site:
if err := openConfig(); err != nil {
    return &ErrHint{
        Err:  err,
        Hint: "create ~/.config/mytool/config.toml or set MYTOOL_CONFIG",
    }
}

// in runner:
var h *ErrHint
if errors.As(err, &h) {
    fmt.Fprintln(os.Stderr, "hint:", h.Hint)
}
```

