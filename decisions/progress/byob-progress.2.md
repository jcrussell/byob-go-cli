---
id: byob-progress.2
title: 'TTY-adaptive: spinner on TTY, rate-limited chatter lines off-TTY'
type: decision
priority: 2
status: open
parent: byob-progress
labels:
  - concurrency
  - iostreams
  - progress
---

## Description

Problem: a spinner on a non-TTY (CI logs, piped stderr) emits ANSI
control characters and `\r` carriage returns that turn the log into
garbage. Suppressing progress entirely off-TTY loses useful signal
for long operations.

Idea: mirror byob-output.1 (TTY-adaptive table printer). The progress
interface has two concrete impls behind it:

- **TTY path:** animated spinner with `\r`-refresh, written to
  `IO.ErrOut`. Colored if `IO.ColorScheme()` allows (byob-iostreams.2).
- **Off-TTY path:** periodic plain-text lines to `IO.ErrOut`,
  rate-limited to one line every ~2 seconds, no ANSI. Uses the same
  chatter channel byob-iostreams.3 already allocated to ErrOut.

Critically, the off-TTY path writes **chatter**, not slog records.
Writing structured logs from progress contradicts byob-logging.4 (logs
default-off). Progress should show something in CI at default log
level, so it uses `fmt.Fprintln(io.ErrOut, ...)` directly.

Tradeoffs: the chatter-line path is one-way — no overwrite, each
line adds to the log. That's fine in CI where logs scroll, and
avoids the ANSI mess. TTY users still get the animated UX.

## Design

The TTY spinner impl lives in byob-progress.3 (library-backed via
`charmbracelet/bubbles/spinner`). This decision owns only the
off-TTY `loggingImpl` — the template's own code for the no-TTY
chatter path.

```go
type loggingImpl struct {
    out      io.Writer
    label    string
    lastEmit atomic.Int64
    minGap   time.Duration // default 2s
}

// Update is best-effort under concurrency: two goroutines hitting
// the rate-limit window simultaneously will see both CAS attempts
// and only one will succeed. The loser silently drops, which is
// the intended dedup behavior — progress is advisory, never
// load-bearing.
func (l *loggingImpl) Update(msg string) {
    now := time.Now().UnixNano()
    last := l.lastEmit.Load()
    if now-last < int64(l.minGap) { // time.Duration is int64 ns
        return
    }
    if !l.lastEmit.CompareAndSwap(last, now) {
        return // another goroutine emitted; skip this one
    }
    fmt.Fprintf(l.out, "%s: %s\n", l.label, msg)
}
```

