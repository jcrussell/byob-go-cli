---
id: byob-testing.2
title: '`go-cmp` for non-trivial test comparisons'
type: byob
priority: 2
status: open
parent: byob-testing
labels:
  - testing
---

## Description

Problem: `==` and `reflect.DeepEqual` are fine for primitives and small
structs, but fall over the moment a test wants to ignore a field, treat
nil and empty slices as equal, normalize a timestamp, or produce a
useful failure message. Hand-rolling field-by-field asserts grows
unmaintainable; rolling your own diff producer is its own project.

Idea: depend on `github.com/google/go-cmp/cmp` for non-trivial
comparisons. `cmp.Diff(want, got, opts...)` returns a directional
`-want +got` string when values differ — empty when they're equal —
and the test owns the failure-message wrapper. `cmpopts` (`EquateEmpty`,
`IgnoreFields`, `IgnoreUnexported`, `SortSlices`, `EquateApproxTime`)
covers the common need-to-relax-comparison cases without manual
normalization. `cmp.Comparer` and `cmp.Transformer` hook in custom
equality without polluting production types with `Equal` methods.

Why not testify (`assert.Equal`, `require.Equal`): testify bundles value
comparison with failure-message generation, locking the assertion shape
to its API. cmp returns a diff string and lets the test author build
the message — composes cleanly with table-driven subtests, error
wrapping, and the `got, want` message ordering the Code Review Comments
wiki names.

Why not `reflect.DeepEqual`: equal-or-not boolean with no diff. Test
output reads "wanted X, got Y" with a giant struct on each side; the
diff between them is left to the reader.

Tradeoffs: third-party dependency, modest API surface to learn. cmp
panics on unexported fields by default — call sites either pass
`cmpopts.IgnoreUnexported` or design fixtures around exported shapes.
Worth the dep; gh CLI, Kubernetes, and most modern Go test codebases
ship it.

When not to use: trivial cases — comparing two ints or two strings.
`if got != "expected" { t.Errorf("Foo() = %q, want %q", got, want) }`
stays clearer than wrapping in cmp.

## Design

```go
import (
    "testing"

    "github.com/google/go-cmp/cmp"
    "github.com/google/go-cmp/cmp/cmpopts"
)

func TestList(t *testing.T) {
    got, err := List(ctx, store, ListOpts{Limit: 10})
    if err != nil {
        t.Fatalf("List(...) error = %v, want nil", err)
    }

    want := []Item{{Name: "alpha"}, {Name: "beta"}}

    if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
        t.Errorf("List(...) mismatch (-want +got):\n%s", diff)
    }
}
```

Pair with table-driven tests: the diff lands inside `t.Run(tc.name, ...)`
and identifies the failing case by name.

