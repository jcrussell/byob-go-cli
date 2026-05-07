---
id: byob-input-validation.5
title: Validate at the Options boundary; fail with FlagErrorf before side effects
type: decision
priority: 2
status: open
parent: byob-input-validation
labels:
  - input-validation
---

## Description

Problem: validation that happens halfway through a runFunc has
already caused damage. The store has been opened, the remote auth
handshake succeeded, a partial write has landed — and *then* the
code discovers `--port` is out of range. Users see "failed to
connect" when the real problem is "your flag value was never going
to work."

Idea: the `Options` struct (byob-command-shape.1) is the boundary between
parsed flags/config and business logic. Put an `Options.Validate()
error` method on it, and call it as the first line of every runFunc
(and the first non-no-op step of `PersistentPreRunE` where
appropriate). Return errors wrapped with `cmdutil.FlagErrorf` so the
top-level runner maps them to exit code 2 (usage) instead of 1
(generic failure).

Declarative flag-group constraints (`MarkFlagsMutuallyExclusive`,
`MarkFlagsOneRequired`, etc.) from byob-command-shape.6 already run before
`RunE`, so they don't need an `Options.Validate()` call. This epic
covers what those helpers can't express: range checks, format
checks, cross-field invariants.

Tradeoffs: duplication when the same field shape appears in multiple
commands. Extract a helper
(`validate.Port(p int) error`) and reuse — same principle as the
`pflag.Value` types from byob-command-shape.7.

## Design

```go
// pkg/cmd/serve/serve.go
type Options struct {
    IO   *iostreams.IOStreams
    Port int
    Bind string
    Cert string
    Key  string
}

func (o *Options) Validate() error {
    if o.Port < 1 || o.Port > 65535 {
        return cmdutil.FlagErrorf("--port must be 1-65535, got %d", o.Port)
    }
    if o.Bind != "" {
        if ip := net.ParseIP(o.Bind); ip == nil {
            return cmdutil.FlagErrorf("--bind must be a valid IP: %q", o.Bind)
        }
    }
    // cross-field invariant: cert and key travel together.
    if (o.Cert == "") != (o.Key == "") {
        return cmdutil.FlagErrorf("--cert and --key must be passed together")
    }
    return nil
}

func serveRun(ctx context.Context, opts *Options) error {
    if err := opts.Validate(); err != nil {
        return err            // FlagErrorf → exit code 2, no side effects yet
    }
    // only now: open sockets, load cert, etc.
    return listen(ctx, opts)
}
```

Reusable field-level helpers live under `internal/validate`:

```go
// internal/validate/port.go
func Port(p int, flag string) error {
    if p < 1 || p > 65535 {
        return cmdutil.FlagErrorf("%s must be 1-65535, got %d", flag, p)
    }
    return nil
}
```

