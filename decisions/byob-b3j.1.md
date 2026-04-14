---
id: byob-b3j.1
title: Inject test doubles through the Factory; never monkey-patch globals
type: decision
priority: 2
status: open
parent: byob-b3j
labels:
- cli
- go
- testing
---

## Description

Problem: tests that capture command output by reassigning
`os.Stdout`, or replace an HTTP client by overwriting a
package-level `var httpClient = ...`, or use `monkey.Patch` to
rewire function pointers — all of these are flaky, non-parallel-safe,
and break the moment another test runs concurrently. They also leak
state between tests and make it impossible to know what the code
under test actually does.

Idea: every swappable dependency on the Factory is either an
interface (Prompter, Store, Browser, HTTP client) or an IOStreams
value. Tests construct a Factory literal with purpose-built fakes
for the interfaces, and with `iostreams.Test()` for the streams.
The command under test gets the test Factory via its constructor.
No globals are touched anywhere.

`iostreams.Test()` returns an IOStreams wired to `bytes.Buffer`
values plus the buffers themselves, so the test can assert on
stdout/stderr contents after running the command. Interface fakes
are tiny (often <50 lines) and can record calls, script canned
responses, or both.

Tradeoffs: you write the fakes. For high-churn or wide interfaces,
code generators (`mockery`, `counterfeiter`) pay off; for narrow
interfaces used by a few tests, hand-rolling stays clearer.
TTY-dependent code paths need explicit `SetStdoutTTY(true)` calls on
the test IOStreams — that's a feature, not a bug (TTY behavior
should be exercised deliberately).

When not to use: never. The principle — "test doubles are
constructor arguments, not package variables" — applies universally.

## Design

```go
// iostreams.Test() returns buffers for output assertions.
func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
    in, out, errOut := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
    return &IOStreams{
        In: in, Out: out, ErrOut: errOut,
    }, in, out, errOut
}

// Interface fakes record calls and script replies.
type fakePrompter struct {
    confirmReply bool
    confirmCalls []string
}
func (f *fakePrompter) Confirm(msg string) (bool, error) {
    f.confirmCalls = append(f.confirmCalls, msg)
    return f.confirmReply, nil
}

// A whole test:
func TestListDefault(t *testing.T) {
    io, _, stdout, stderr := iostreams.Test()
    p := &fakePrompter{confirmReply: true}
    f := &Factory{IOStreams: io, Prompter: p, Store: fakeStore}

    cmd := NewCmdList(f, nil)
    cmd.SetArgs([]string{"--format", "tsv"})
    require.NoError(t, cmd.Execute())
    require.Contains(t, stdout.String(), "item-1\titem-2")
    require.Empty(t, stderr.String())
}
```

