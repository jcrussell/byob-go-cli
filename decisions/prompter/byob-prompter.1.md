---
id: byob-prompter.1
title: Prompter as a narrow consumer interface
type: decision
priority: 2
status: open
parent: byob-prompter
labels:
  - context
  - interfaces
  - prompter
---

## Description

Problem: a wide prompter interface (`AskOne`, `AskMany`, 20 options
per call) couples every command to a specific library's question
shape. When the library changes — or the implementation swaps for
tests — the blast radius is large.

Idea: five methods is the minimum that covers a real CLI. Each
takes `context.Context` as its first argument so prompts inherit
the caller's deadline and cancellation — a command with a deadline
should be able to abandon a hung prompt instead of waiting on
stdin forever:

```go
type Prompter interface {
    Confirm(ctx context.Context, msg string, def bool) (bool, error)
    Input(ctx context.Context, msg, def string) (string, error)
    Password(ctx context.Context, msg string) (string, error)
    Select(ctx context.Context, msg string, options []string) (int, error)
    MultiSelect(ctx context.Context, msg string, options []string) ([]int, error)
}
```

Per byob-interfaces.1, the interface lives in `pkg/cmd/prompt/` (or closer
to the consumer) and concrete impls (live, stub) satisfy it
structurally. No library type leaks into the consumer.

Tradeoffs: five methods won't cover every fancy UX (autocomplete
pickers, path completers, multi-line editors). When you need one,
add a dedicated method — don't stretch `Input` to do it. The ctx
parameter is uniform across all methods even though most callers
will pass the command's root ctx untouched — the consistency
matters more than per-method ergonomics.

## Design

```go
// pkg/cmd/prompt/prompt.go
package prompt

type Prompter interface {
    Confirm(ctx context.Context, msg string, def bool) (bool, error)
    Input(ctx context.Context, msg, def string) (string, error)
    Password(ctx context.Context, msg string) (string, error)
    Select(ctx context.Context, msg string, options []string) (int, error)
    MultiSelect(ctx context.Context, msg string, options []string) ([]int, error)
}

// Factory holds it as an eager field (cheap, like IOStreams):
type Factory struct {
    IOStreams *iostreams.IOStreams
    Prompter  prompt.Prompter
    // ...lazy fields
}
```

Commands accept `prompt.Prompter` as an `Options` field, not a
library type. Tests pass `prompt.Stub{...}`; production passes
`prompt.NewLive(io)`.

