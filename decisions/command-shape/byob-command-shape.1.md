---
id: byob-command-shape.1
title: Three-part command shape with a runF test-injection hook
type: byob
priority: 2
status: open
parent: byob-command-shape
labels:
  - command-shape
---

## Description

Problem: a cobra command that parses flags, opens resources, and executes
business logic in one `RunE` function is untestable and unreadable. Even
once you split it, testing the *parsing* without executing the business
logic still wants a seam — otherwise every flag-parsing test sets up
(or mocks) every dependency and asserts on side effects, which is
integration-test territory.

Idea: split every subcommand into three pieces.
(1) `Options` struct — holds dependencies (pulled from Factory) and parsed
flag values.
(2) `NewCmdXxx(f *Factory, runF func(*Options) error) *cobra.Command` —
binds flags, constructs Options, wires `RunE`. The `runF` parameter is a
test-injection hook: inside `RunE`, `if runF != nil { return runF(opts) }`
takes the test path. Production code passes `nil`; tests pass a closure
that captures the parsed Options.
(3) A package-private `xxxRun(opts *Options) error` — pure business logic.

Each layer is independently testable. Flag-parsing tests call the
constructor with a `runF` that captures Options and returns nil — no real
work executes. Business-logic tests call `xxxRun` with a handcrafted
Options.

Tradeoffs: three functions per command instead of one, plus an extra
parameter on every constructor. You give up the ability to export the
runFunc directly. Small file tax that pays back immediately when you
write the first test.

## Design

```go
type Options struct {
    IO    *iostreams.IOStreams
    Store func() (Store, error)
    Name  string
    Force bool
}

func NewCmdCreate(f *Factory, runF func(*Options) error) *cobra.Command {
    opts := &Options{IO: f.IOStreams, Store: f.Store}
    cmd := &cobra.Command{
        Use:  "create <name>",
        Args: cobra.ExactArgs(1),
        RunE: func(c *cobra.Command, args []string) error {
            opts.Name = args[0]
            if runF != nil { return runF(opts) }
            return createRun(opts)
        },
    }
    cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "overwrite existing")
    return cmd
}

func createRun(opts *Options) error { /* pure business logic */ }
```

Production wiring vs. test wiring:

```go
// production:
cmd := NewCmdCreate(f, nil)

// test:
var got *Options
cmd := NewCmdCreate(f, func(o *Options) error { got = o; return nil })
cmd.SetArgs([]string{"myname", "--force"})
err := cmd.Execute()

require.NoError(t, err)
require.Equal(t, "myname", got.Name)
require.True(t, got.Force)
```

