---
id: byob-progress.3
title: 'Library pick: bubbles/spinner + schollz/progressbar, wrapped'
type: byob
priority: 2
status: open
parent: byob-progress
labels:
  - deps-philosophy
  - progress
---

## Description

Problem: rendering a spinner or a progress bar portably across
Windows, macOS, and Linux terminals is more code than it looks — ANSI
cursor movement, terminal-width detection, `\r`-handling when the
terminal scrolls, SIGWINCH response. Writing it in-house is possible
but fiddly and off the tool's main value path.

Idea: wrap two libraries behind the Progress interface (byob-progress.1):

- **`charmbracelet/bubbles/spinner`** for unknown-total operations.
  It's a bubbletea component, so the impl runs a tiny `tea.Program`
  and drives it via messages. The transitive weight (bubbletea +
  lipgloss) is already paid by byob-prompter.2's huh pick, so the
  marginal cost here is a direct import line, not a new dep tree.
  In exchange: a maintained, themeable spinner aligned with the
  same rendering stack the prompter uses.
- **`schollz/progressbar/v3`** for known-total operations. Pure-Go,
  actively maintained, drop-in for the fetch-and-iterate case —
  kept standalone to avoid forcing every caller to reason about a
  bubbletea `Program` for trivial progress.

Neither type escapes the `progress` package; callers see the
`Progress` interface only. This preserves the option to replace the
impl later (e.g. consolidate both paths on `bubbles/progress` when
the tool grows a richer TUI).

When to graduate further: when the tool acquires a non-trivial TUI
(dashboard, multi-pane, form-driven flows), consolidate the
known-total path on `bubbles/progress` too and drop schollz. Until
then, keeping a dedicated progress-bar impl avoids bubbletea
ceremony on the hot fetch path.

Tradeoffs:

- **Bubbletea ceremony for the spinner.** bubbles/spinner is not a
  `.Start()`/`.Stop()` wrapper — the impl has an `Init`/`Update`/
  `View` model and a goroutine running `tea.Program.Run()`.
  Heavier than a procedural spinner, but worth it for the
  maintenance story and the alignment with byob-prompter.2's rendering
  stack.
- **Prior pick (`briandowns/spinner`) was tiny but dormant.** This
  decision previously wrapped it on a "small, stable API"
  argument; upstream activity has dropped and the low-cost
  argument stopped outweighing the bitrot risk. The swap to
  bubbles accepts more ceremony in exchange for an active
  maintainer.
- **Pure-Go budget intact.** Both bubbletea and schollz honor
  byob-release.8 (no CGO).

## Design

```go
// pkg/cmd/progress/spinner.go
package progress

import (
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
)

type spinnerImpl struct {
    prog *tea.Program
}

type spinnerModel struct {
    sp    spinner.Model
    label string
    final string
    done  bool
}

type updateLabelMsg string
type stopMsg struct{ final string }

func (m *spinnerModel) Init() tea.Cmd { return m.sp.Tick }
func (m *spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case updateLabelMsg:
        m.label = string(msg)
    case stopMsg:
        m.final = msg.final
        m.done = true
        return m, tea.Quit
    case spinner.TickMsg:
        var cmd tea.Cmd
        m.sp, cmd = m.sp.Update(msg)
        return m, cmd
    }
    return m, nil
}
func (m *spinnerModel) View() string {
    if m.done { return m.final }
    return m.sp.View() + " " + m.label
}

func NewSpinner(ctx context.Context, out io.Writer, label string) Progress {
    sp := spinner.New()
    sp.Spinner = spinner.Dot
    m := &spinnerModel{sp: sp, label: label}
    // tea.WithContext causes Program.Run() to return when ctx is
    // cancelled — no separate watcher goroutine needed.
    prog := tea.NewProgram(m, tea.WithOutput(out), tea.WithContext(ctx))
    return &spinnerImpl{prog: prog}
}
func (p *spinnerImpl) Start()            { go func() { _, _ = p.prog.Run() }() }
func (p *spinnerImpl) Update(msg string) { p.prog.Send(updateLabelMsg(msg)) }
func (p *spinnerImpl) Stop()             { p.prog.Send(stopMsg{}) }
func (p *spinnerImpl) Fail(err error)    { p.prog.Send(stopMsg{final: "✗ " + err.Error() + "\n"}) }

// pkg/cmd/progress/bar.go uses schollz/progressbar/v3 similarly
// (standalone, not bubbletea — see tradeoffs above).
```

