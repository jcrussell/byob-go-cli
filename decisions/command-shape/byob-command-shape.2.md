---
id: byob-command-shape.2
title: runF test-injection hook inside RunE
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

Problem: testing a cobra command that actually calls its business logic
requires setting up (or mocking) every dependency and asserting on side
effects. That's integration-test territory; you want unit tests for flag
parsing.

Idea: every `NewCmdXxx` takes an optional `runF func(*Options) error`
parameter. Inside `RunE`, `if runF != nil { return runF(opts) }` takes the
test path. Production code always passes `nil`. Tests pass a closure that
captures the parsed Options and returns nil.

Tradeoffs: one extra parameter on every command constructor. You give up the
ability to export the runFunc directly, but you gain clean unit tests for
flag parsing, argument handling, and validation without any real work
executing.

## Design

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

