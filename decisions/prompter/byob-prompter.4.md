---
id: byob-prompter.4
title: Test prompter is a scripted FIFO stub
type: byob
priority: 2
status: open
parent: byob-prompter
labels:
  - prompter
  - testing
---

## Description

Problem: testing code that prompts the user typically involves
fragile input-simulation (writing to a pipe, racing a goroutine) or
a mocking library that hides the call sequence.

Idea: a `Stub` type that holds per-method FIFOs. Each call pops the
next value. Over-consumption panics (loud failure beats silent
incorrect answer). Ordering matches the call order in the test — if
the test expects "Confirm, then Input, then Select", the stub's
slices are sized accordingly.

Aligns with byob-testing.1 (test doubles through the Factory, no
monkey-patching) — the stub replaces `f.Prompter` directly; no
library-level patching is needed.

Tradeoffs: when a command changes the order of its prompts, the
test slice doesn't auto-resize — you update the slice, which is
exactly the right level of refactor pain. For commands that branch
the prompt order based on prior answers, the stub's per-method
FIFOs are not order-preserving across methods; write a record-calls
helper if that matters.

## Design

```go
// pkg/cmd/prompt/stub.go
package prompt

type Stub struct {
    Confirms     []bool
    Inputs       []string
    Passwords    []string
    Selects      []int
    MultiSelects [][]int
}

func (s *Stub) Confirm(_ context.Context, msg string, def bool) (bool, error) {
    if len(s.Confirms) == 0 {
        panic("prompt.Stub: no Confirms left for " + msg)
    }
    v := s.Confirms[0]; s.Confirms = s.Confirms[1:]
    return v, nil
}
// Input/Password/Select/MultiSelect identical shape — each takes
// context.Context as the first param to satisfy the interface
// (byob-prompter.1) and ignores it; tests don't cancel mid-call.
```

Used in tests:

```go
f := &Factory{
    IOStreams: iostreams.Test(),
    Prompter:  &prompt.Stub{Confirms: []bool{true}, Inputs: []string{"foo"}},
}
```

