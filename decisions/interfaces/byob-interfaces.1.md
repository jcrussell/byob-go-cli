---
id: byob-interfaces.1
title: Define interfaces in the consumer package, narrow to what's used
type: decision
priority: 2
status: open
parent: byob-interfaces
labels:
  - interfaces
---

## Description

Problem: commands that depend on concrete types can't be tested without
mocking deep internals, and they can't be extended to a second backend
without editing the command.

Idea: define interfaces in the *consumer* package, narrow to what the
consumer actually uses. Concrete implementations live elsewhere and satisfy
the interface structurally. "Accept interfaces, return structs" — at every
package seam that matters.

Common seams:
- `Store` — commands use `List`, `Get`, `Save`; implementations are sqlite,
  postgres, in-memory.
- `Prompter` — commands use `Confirm`, `Select`; implementations are live
  (stdin-reading) and fake (scripted replies).
- `Backend` — commands use `Create`, `Destroy`; implementations are libvirt,
  proxmox, ...

Tradeoffs: slightly more types. Huge payoff the first time you add a second
backend or write a test that needs a fake.

## Design

```go
// pkg/cmd/list/list.go
type listStore interface {
    ListItems(ctx context.Context) ([]Item, error)
}

type Options struct {
    IO    *iostreams.IOStreams
    Store func() (listStore, error) // narrow!
}

// pkg/cmd/list/list_test.go
type fakeStore struct{ items []Item }
func (f *fakeStore) ListItems(context.Context) ([]Item, error) {
    return f.items, nil
}

// The production sqlite.Store satisfies both listStore and createStore
// structurally, without importing either cmd package.
```

