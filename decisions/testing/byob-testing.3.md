---
id: byob-testing.3
title: Assert on behavior, not implementation
type: decision
priority: 2
status: open
parent: byob-testing
labels:
  - testing
---

## Description

Problem: tests that assert on internal call counts ("expected
`Store.Get` to be called exactly twice"), private field values, or the
precise order of collaborator interactions break the moment the
implementation refactors — even when the user-visible behavior is
unchanged. The test becomes a snapshot of one specific implementation,
not a contract about what the code does. Maintenance cost compounds:
every internal refactor triggers test churn that adds no signal, and
the suite slowly trains its authors to treat test failures as noise.

Idea: assert on the **behavior** the caller can observe — return
values, errors (including their `errors.Is`/`errors.As` shape), and
the side effects the code is responsible for. For commands, that means
the bytes in the stdout/stderr buffers from `iostreams.Test()` (per
byob-testing.1), the exit error, and any persisted state the test set
up the storage layer to inspect. It does *not* mean the sequence of
method calls on a `fakePrompter`, the internal cursor position of a
paginator, or whether `Store.Get` was called before `Store.List`.

The fakes from byob-testing.1 make this easy to get wrong — once a
fake records calls, the temptation is to assert on those records.
Resist except where the call itself *is* the behavior under test (e.g.
"the dry-run mode does not invoke `Store.Write`"). Treat recorded
calls as a debugging aid, not the default assertion surface.

Tradeoffs: behavior-focused tests can miss regressions in internal
state that don't surface through the public boundary. Mitigation: make
the boundary wider when it matters — return the new state, expose an
inspection method, emit a log line the test can assert on. If you
find yourself wanting to test a private invariant, the invariant
probably wants to be observable.

When not to use: characterization tests written specifically to lock
in current implementation behavior before a refactor (rare, and
delete them after the refactor). Tests that assert on call counts as
a proxy for algorithmic complexity (uncommon in CLIs).

## Design

```go
// Behavior-focused: assert on what the user sees.
func TestListPrintsItems(t *testing.T) {
    io, _, stdout, _ := iostreams.Test()
    f := &Factory{IOStreams: io, Store: fakeStoreWithItems("alpha", "beta")}

    cmd := NewCmdList(f, nil)
    require.NoError(t, cmd.Execute())

    if diff := cmp.Diff("alpha\nbeta\n", stdout.String()); diff != "" {
        t.Errorf("stdout mismatch (-want +got):\n%s", diff)
    }
}

// Implementation-focused: brittle. The test now owns the call
// sequence, not the behavior. Refactor `List` to cache or batch, and
// this breaks without any user-visible change.
func TestListCallsStoreGetTwice(t *testing.T) {
    s := &recordingStore{}
    cmd := NewCmdList(&Factory{Store: s}, nil)
    _ = cmd.Execute()

    if len(s.getCalls) != 2 {
        t.Errorf("Store.Get calls = %d, want 2", len(s.getCalls))
    }
}
```

The exception — when the *absence* of a call is the contract:

```go
func TestDryRunDoesNotWrite(t *testing.T) {
    s := &recordingStore{}
    cmd := NewCmdDelete(&Factory{Store: s}, nil)
    cmd.SetArgs([]string{"--dry-run", "item-1"})
    require.NoError(t, cmd.Execute())

    if len(s.writeCalls) != 0 {
        t.Errorf("Store.Write called %d times in dry-run, want 0",
            len(s.writeCalls))
    }
}
```

Here the no-op behavior *is* the user-visible contract of `--dry-run`,
so asserting on the absence of the call is the same as asserting on
behavior. The shape is the exception, not the default.
