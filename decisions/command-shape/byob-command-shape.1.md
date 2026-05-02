---
id: byob-command-shape.1
title: 'Three-part command shape: Options + NewCmdXxx(f, runF) + runFunc'
type: decision
priority: 2
status: open
parent: byob-command-shape
labels:
  - cli
  - command-shape
  - go
---

## Description

Problem: a cobra command that parses flags, opens resources, and executes
business logic in one `RunE` function is untestable and unreadable.

Idea: split every subcommand into three pieces.
(1) `Options` struct — holds dependencies (pulled from Factory) and parsed
flag values.
(2) `NewCmdXxx(f *Factory, runF func(*Options) error) *cobra.Command` —
binds flags, constructs Options, wires `RunE`.
(3) A package-private `xxxRun(opts *Options) error` — pure business logic.

Each layer is independently testable. Flag parsing tests call the constructor
with a mock `runF`. Business logic tests call `xxxRun` with a handcrafted
Options.

Tradeoffs: three functions per command instead of one. A small file tax that
pays back immediately when you write the first test.

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

