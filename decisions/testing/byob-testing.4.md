---
id: byob-testing.4
title: Tests land with the code, not after
type: decision
priority: 2
status: open
parent: byob-testing
labels:
  - testing
---

## Description

Problem: tests deferred to "I'll add them once the feature works"
rarely arrive in the same shape, and often don't arrive at all. The
implementation hardens around an untestable surface — package-level
state, awkward signatures, hidden side effects — and the eventual test
backfill either reshapes the production code under pressure or settles
for the few assertions the existing API happens to permit. Either
way, the suite becomes a partial witness of the codebase rather than a
definition of its contract.

Idea: every meaningful change ships with its tests in the same commit
(or in the immediately preceding commit, if the workflow is test-first).
The byob command shape is built to make this cheap: the `runF`
injection hook on the Options struct (byob-command-shape.1) lets the
test construct an Options literal and call `runF` directly, without
invoking cobra. There is no "I'll add tests once the command is wired
up" excuse, because the testable boundary exists from line one.

The cadence in practice:

1. Sketch the Options struct and the `runF` signature.
2. Write the first test against `runF` (often a failing one).
3. Implement `runF` until the test passes.
4. Add the cobra wrapper (`NewCmdXxx`) and a single integration test
   that exercises the full path through `cmd.Execute()`.
5. Run `make test` (and `make lint`, per byob-release.9 — same
   pre-commit gate, both checks) before the commit lands.

This is not dogmatic TDD — the order of steps 2 and 3 can flip, and it
often does. The non-negotiable is that step 5 runs every time, and
that the diff under review always contains both production code and
tests.

Tradeoffs: upfront friction. Naming an Options struct and a `runF`
before the implementation is fully sketched can feel premature, and
sometimes the API does change after the test is written. That's a
feature: the friction surfaces shape problems at design time, not
after the code has solidified. The cost is small for byob-shaped
commands because the boilerplate is mechanical (byob-command-shape.1
walks through it).

When not to use: spikes and throwaway experiments — but mark them as
such, branch them, and don't merge them. The "I'm just exploring"
exemption decays into "this is production code with no tests" faster
than expected.

## Design

The mechanical cadence on a new subcommand:

```go
// 1. Options + runF sketch (no real implementation yet).
type ListOptions struct {
    IO    *iostreams.IOStreams
    Store Store

    Format string

    runF func(*ListOptions) error
}

func listRun(opts *ListOptions) error { return nil } // placeholder

// 2. First test — fails until listRun is real.
func TestListRunTSV(t *testing.T) {
    io, _, stdout, _ := iostreams.Test()
    opts := &ListOptions{
        IO:     io,
        Store:  fakeStoreWithItems("alpha", "beta"),
        Format: "tsv",
    }
    if err := listRun(opts); err != nil {
        t.Fatalf("listRun() error = %v, want nil", err)
    }
    if got, want := stdout.String(), "alpha\tbeta\n"; got != want {
        t.Errorf("listRun() stdout = %q, want %q", got, want)
    }
}

// 3. Implement listRun until the test passes.
// 4. Wrap in NewCmdList; add one cmd.Execute() integration test.
// 5. make test before commit.
```

The `runF` field on Options is what makes step 2 cheap. Without it,
every test would route through cobra (`cmd.Execute()`), which is fine
for the integration test in step 4 but heavyweight for the unit tests
in step 2.

