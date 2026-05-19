---
id: byob-iostreams.1
title: IOStreams wraps In/Out/ErrOut + TTY flags; commands never touch os.Std*
type: byob
priority: 2
status: open
parent: byob-iostreams
labels:
  - iostreams
---

## Description

Problem: `fmt.Println`, `os.Stdout.Write`, and friends scattered across a
codebase are impossible to capture in tests and impossible to redirect
consistently (e.g., to suppress output in a JSON mode).

Idea: define a small `IOStreams` struct with `In io.Reader`, `Out io.Writer`,
`ErrOut io.Writer`, plus TTY flags for all three streams. Every command
writes through its `*IOStreams`. The only place in the codebase that
touches `os.Stdin` / `os.Stdout` / `os.Stderr` is `iostreams.System()`,
called once in `main()`.

All three streams carry their own TTY flag. Stdin's flag isn't just
symmetry: the Prompter (byob-prompter) consults `IsStdinTTY()` to decide
whether to prompt at all, so this decision has to expose it alongside
the output-stream flags.

Tradeoffs: you pay a tiny indirection. You also gain the ability to swap
buffers in tests, suppress chatter in JSON mode, and ask `IsStdoutTTY()`
without sprinkling `isatty` checks across the codebase.

## Design

```go
type IOStreams struct {
    In     io.Reader
    Out    io.Writer
    ErrOut io.Writer

    stdinIsTTY  bool
    stdoutIsTTY bool
    stderrIsTTY bool
    colorScheme *ColorScheme
}

func (s *IOStreams) IsStdinTTY() bool  { return s.stdinIsTTY }
func (s *IOStreams) IsStdoutTTY() bool { return s.stdoutIsTTY }
func (s *IOStreams) IsStderrTTY() bool { return s.stderrIsTTY }
func (s *IOStreams) ColorScheme() *ColorScheme { return s.colorScheme }

func System() *IOStreams {
    return &IOStreams{
        In:  os.Stdin,
        Out: os.Stdout,
        ErrOut: os.Stderr,
        stdinIsTTY:  isatty.IsTerminal(os.Stdin.Fd()),
        stdoutIsTTY: isatty.IsTerminal(os.Stdout.Fd()),
        stderrIsTTY: isatty.IsTerminal(os.Stderr.Fd()),
    }
}
```

