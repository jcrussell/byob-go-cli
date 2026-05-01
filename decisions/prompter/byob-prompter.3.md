---
id: byob-prompter.3
title: Refuse to prompt when stdin is not a TTY
type: decision
priority: 2
status: open
parent: byob-prompter
labels:
- cli
- go
- iostreams
- prompter
---

## Description

Problem: a command that prompts for confirmation works interactively
but hangs in CI or in a pipeline with no stdin attached. Worse,
naive prompt libraries read EOF, interpret it as "no", and silently
skip a destructive action without failing.

Idea: every Prompter method checks `IO.IsStdinTTY()` first (byob-iostreams.1
exposes this on IOStreams). If false, return a sentinel `ErrNotTTY`.
The caller is responsible for handling it — typically by requiring
`--yes`/`-y` (see byob-prompter.5) or failing with a clear message
("pass --yes to skip confirmation in non-interactive environments").

Tradeoffs: every Prompter caller now has to think about the non-TTY
case. That's the point — the failure mode goes from "silently
wrong" to "explicitly unsupported." `--yes` (byob-prompter.5) already
covers the "skip prompts despite a TTY" case; no separate env-var
escape hatch is needed.

## Design

```go
var ErrNotTTY = errors.New("no TTY available for prompting")

func (p *live) Confirm(ctx context.Context, msg string, def bool) (bool, error) {
    if !p.io.IsStdinTTY() {
        return false, ErrNotTTY
    }
    // ...prompt (see byob-prompter.2 for the ctx-select pattern)
}

// Caller:
yes, err := f.Prompter.Confirm(ctx, "Delete all?", false)
switch {
case errors.Is(err, prompt.ErrNotTTY):
    return fmt.Errorf("not a TTY; pass --yes to confirm non-interactively")
case err != nil:
    return err
case !yes:
    return ErrCancel // see byob-errors.1
}
```

Pairs with the `prompter-tty-check` memory — same rule, one-line
form, always-on.

