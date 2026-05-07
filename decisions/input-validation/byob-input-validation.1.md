---
id: byob-input-validation.1
title: Resolve and containment-check every user-supplied path
type: decision
priority: 2
status: open
parent: byob-input-validation
labels:
  - input-validation
---

## Description

Problem: a user path like `../../etc/passwd` or a symlink like
`safe.txt -> /etc/passwd` lets a caller read or write outside the
directory the tool expected to operate in. `filepath.Clean` alone is
not sufficient — it collapses `../` segments but says nothing about
symlinks, and an uncleaned absolute path (`/etc/passwd`) bypasses it
entirely.

Idea: two steps, in this order, at every entry point where a path
crosses the trust boundary (flag value, config value, CLI arg):

1. **Resolve.** `path, err := filepath.EvalSymlinks(filepath.Join(base, input))`
   — follows symlinks and normalizes.
2. **Containment-check.** `rel, err := filepath.Rel(base, path);
   if err != nil || strings.HasPrefix(rel, "..") { reject }`. A
   relative path that starts with `..` after `Rel` means the target
   escaped `base`.

Wrap both in a helper (`safejoin.Resolve(base, input)`) and use it
exclusively. Reject with a `FlagErrorf` (byob-errors.1) so the top-level
runner maps to exit 2.

Tradeoffs: `EvalSymlinks` requires the target to exist at resolve
time. For "create this file inside base" flows, resolve the *parent
directory* instead and re-check containment. Go 1.24 also adds
`os.Root` which is stricter (and cleaner) for a pure-filesystem
sandbox — use that where applicable.

When not to use: paths the user never controls (constants, embedded
assets). The seam is untrusted input, not paths in general.

## Design

```go
// internal/safejoin/safejoin.go
package safejoin

import (
    "errors"
    "path/filepath"
    "strings"
)

var ErrEscapesBase = errors.New("path escapes allowed base directory")

// Resolve joins input onto base, follows symlinks, and refuses to
// return a path outside base.
func Resolve(base, input string) (string, error) {
    joined := filepath.Join(base, input)
    resolved, err := filepath.EvalSymlinks(joined)
    if err != nil {
        return "", err
    }
    rel, err := filepath.Rel(base, resolved)
    if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
        return "", ErrEscapesBase
    }
    return resolved, nil
}
```

For Go 1.24+, prefer `os.Root` for the whole traversal:

```go
root, err := os.OpenRoot(base)   // refuses to escape
if err != nil { return err }
defer root.Close()
f, err := root.Open(input)       // any ../ or absolute path fails here
```

