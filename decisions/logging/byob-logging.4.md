---
id: byob-logging.4
title: Logs and chatter share ErrOut; quiet default keeps them separate
type: decision
priority: 2
status: open
parent: byob-logging
labels:
  - cli
  - go
  - iostreams
  - logging
---

## Description

Problem: byob-iostreams.3 allocated `ErrOut` to human chatter (progress,
prompts, warnings). A slog handler wired to the same stream can
interleave structured log records into "Loading items…" chatter —
now `2>tool.log` captures garbage and a scripter piping stderr to
`jq` sees prose mixed with JSON.

Idea: the default log level is `Warn` (byob-logging.3), so in the common
case logs emit nothing and chatter owns ErrOut by itself. When the
user passes `-v`/`-vv`/`--log-level`, they've explicitly opted into a
mixed stream — they know what they asked for. For scripted capture
of structured output, `--log-format=json` swaps `TextHandler` to
`JSONHandler`. For full separation, `--log-file=<path>` writes logs
to a file, leaving ErrOut to chatter alone.

Rule of thumb: if the user hasn't asked for logs, don't print any.
Chatter (byob-iostreams.3) is still the primary human channel. Logs are
machine-parseable opt-in.

No `--quiet`/`-q` flag. The template deliberately omits one because
the byob-iostreams.3 / byob-logging.4 combination already gives users three
orthogonal knobs: `2>/dev/null` silences ErrOut entirely,
`--log-file=/dev/null` discards logs while keeping chatter, and the
default (no flags) is already quiet for logs. Adding `--quiet` would
overlap confusingly with these. Projects that really want one can
add a persistent `-q` that redirects `IO.ErrOut` to `io.Discard`
in `PersistentPreRunE` — that's the one-line answer.

Tradeoffs: "quiet by default" means new developers don't see logs
until they know about `-v`. That's fine — the chatter channel still
gives them what they need, and `-v` is the standard debugging
reflex for Unix tools.

The `--log-file` file handle deliberately lives for the process
lifetime; every `slog` call would need a reference otherwise. A
short-lived CLI exits and the OS reclaims it. If a tool later grows a
long-running / daemon mode, wire the close into the root command's
shutdown hook rather than deferring it next to the open.

## Design

```go
// opts here is the root command's options struct (LogFile, LogFormat, etc.).
// The slog handler options get their own name to avoid shadowing.
var w io.Writer = f.IOStreams.ErrOut
if opts.LogFile != "" {
    fh, err := os.OpenFile(opts.LogFile,
        os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil { return err }
    w = fh
}

hopts := &slog.HandlerOptions{Level: lvl}
var h slog.Handler
switch opts.LogFormat {
case "json":
    h = slog.NewJSONHandler(w, hopts)
default:
    h = slog.NewTextHandler(w, hopts)
}
logger := slog.New(h)
```

