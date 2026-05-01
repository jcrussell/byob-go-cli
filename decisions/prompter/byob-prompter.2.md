---
id: byob-prompter.2
title: 'Library pick: charmbracelet/huh for the live prompter'
type: decision
priority: 2
status: open
parent: byob-prompter
labels:
- cli
- deps-philosophy
- go
- prompter
---

## Description

Problem: every prompt library has trade-offs. `AlecAivazis/survey/v2`
is in maintenance mode; `charmbracelet/huh` is actively maintained
and pretty; raw `bubbletea` is too low-level to be a prompter. A
stdlib-only live prompter is doable (~120 lines + `golang.org/x/term`
for no-echo passwords) but ships an underwhelming Select/MultiSelect
UX compared to an arrow-key picker.

Idea: ship `charmbracelet/huh` as the default `prompt.NewLive` impl.
It covers Confirm / Input / Password / Select / MultiSelect with a
consistent TTY UX, degrades deterministically when stdin isn't a
terminal (coordinated with byob-prompter.3's `ErrNotTTY` sentinel), and
is actively maintained by a team that ships other CLI-adjacent
libraries.

Because the Prompter is a narrow interface (byob-prompter.1), huh's types
never leak to callers; swapping the impl later is mechanical.

Tradeoffs:

- **Transitive dep weight.** huh pulls bubbletea + lipgloss. Against
  byob-dependencies.1's pure-Go minimalism the UX win is worth the weight,
  but it IS a meaningful ask — flag it explicitly in `go.mod`
  review.
- **Known risk: API churn.** Charmbracelet's ecosystem has shipped
  breaking changes across minor/major revisions more than once.
  The narrow interface insulates callers, but the `live` impl in
  this file will need rework on those bumps. Treat the `huh` pin
  in `go.mod` as a conscious version choice, not a floating
  dependency, and expect to spend time on updates.
- **Known risk: context cancellation is best-effort.** `huh.Run()`
  is blocking and has no native `context.Context` awareness, so
  the `live` impl runs huh in a goroutine and selects on
  `ctx.Done()`. On cancellation the prompt goroutine leaks until
  the user hits a key (or the process exits); the *caller* gets
  `ctx.Err()` immediately, which is what matters for command
  shutdown.
- **Fallback if the above risks bite.** Switch to a stdlib-only
  impl (`bufio.Scanner` + `golang.org/x/term` for no-echo
  passwords) or `AlecAivazis/survey/v2` (stable API, maintenance-
  mode but functional). Both satisfy the same `Prompter`
  interface.

## Design

```go
// pkg/cmd/prompt/live.go
package prompt

import "github.com/charmbracelet/huh"

type live struct { io *iostreams.IOStreams }

func NewLive(io *iostreams.IOStreams) Prompter { return &live{io: io} }

func (p *live) Confirm(ctx context.Context, msg string, def bool) (bool, error) {
    if !p.io.IsStdinTTY() { return false, ErrNotTTY }
    v := def // huh reads the initial value as the default selection
    errCh := make(chan error, 1)
    go func() {
        errCh <- huh.NewConfirm().
            Title(msg).
            Affirmative("Yes").
            Negative("No").
            Value(&v).
            WithTheme(huh.ThemeBase()).
            Run()
    }()
    select {
    case err := <-errCh:
        if err != nil { return false, err }
        return v, nil
    case <-ctx.Done():
        // huh has no ctx hook; the goroutine keeps running until the
        // user hits a key or the process exits. Caller gets ctx.Err()
        // immediately — that's the contract that matters for shutdown.
        return false, ctx.Err()
    }
}

func (p *live) Select(ctx context.Context, msg string, options []string) (int, error) {
    if !p.io.IsStdinTTY() { return 0, ErrNotTTY }
    opts := make([]huh.Option[int], len(options))
    for i, o := range options { opts[i] = huh.NewOption(o, i) }
    var v int
    errCh := make(chan error, 1)
    go func() {
        errCh <- huh.NewSelect[int]().Title(msg).Options(opts...).Value(&v).Run()
    }()
    select {
    case err := <-errCh:
        return v, err
    case <-ctx.Done():
        return 0, ctx.Err()
    }
}

// Input/Password/MultiSelect follow the same ctx-select shape.
```

Tests never instantiate `live` — they use `prompt.Stub` (byob-prompter.4),
so huh never runs in the test path.

