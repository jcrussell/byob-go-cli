---
id: byob-release.10
title: Stdlib-first; reach for deps only when the stdlib is genuinely insufficient
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - deps-philosophy
  - release
---

## Description

Problem: every direct dependency is a permanent maintenance commitment
— security updates, breaking changes, transitive bloat, the risk that
an upstream goes unmaintained. The Go standard library covers most of
what a CLI needs (`net/http`, `encoding/json`, `os/exec`, `log/slog`,
`database/sql`, `context`, `flag`/`pflag`), and the gap between
"writing it myself with stdlib" and "pulling a wrapper library" is
usually a hundred lines that the wrapper would have written anyway —
but without the wrapper's API surface, version pinning, and
supply-chain risk.

Idea: the default is the standard library. Reach for an external
dependency only when the stdlib is genuinely insufficient *and* the
dep itself follows pure-Go discipline (byob-release.8). The two
decisions stack: byob-release.8 is the floor any accepted dep must
clear (pure-Go); this decision is the prior question of whether to
take a dep at all. The blessed exceptions byob ships are:

- `github.com/spf13/cobra` — the CLI command tree (byob-command-shape).
  Stdlib `flag` doesn't do subcommand trees.
- `github.com/google/go-cmp` — non-trivial test diffs
  (byob-testing.2). Stdlib `reflect.DeepEqual` returns a boolean with
  no diff.
- `modernc.org/sqlite` — pure-Go SQLite driver (byob-release.8). No
  stdlib equivalent.
- `github.com/goreleaser/goreleaser` — release matrix (build-time
  only, not a runtime dep) (byob-release.4).

Each exception has a decision bead that justifies it. A new dependency
proposal needs the same form: name the stdlib API it replaces, explain
why that API is insufficient *for this CLI*, and confirm the dep is
pure-Go.

Tradeoffs: writing it yourself takes longer than `go get` and pasting
a snippet from the README. The payoff is a smaller `go.sum`, faster
builds, fewer security-advisory false alarms, and a codebase that
survives upstream churn. For a CLI whose lifetime is measured in
years, the math favors stdlib-first by a wide margin.

When not to use: prototypes and spikes. If you're sketching to learn
whether an idea works at all, grab the dep. Before the prototype
becomes the real thing, re-evaluate every dep against the stdlib-first
test below.

## Design

A worked example. The HTTP client (byob-http-client) is `net/http.Client`
with a configured `Transport` — not `resty`, not `req`, not
`gentleman`. The stdlib gives you:

```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        IdleConnTimeout:     90 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
    },
}

req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
if err != nil {
    return fmt.Errorf("building request: %w", err)
}
req.Header.Set("User-Agent", userAgent)

resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("requesting %s %s: %w", req.Method, req.URL, err)
}
defer resp.Body.Close()
```

That's the entire HTTP layer for a typical CLI. Wrappers add: a fluent
builder, automatic retries (which you usually want explicit anyway),
automatic JSON marshalling (one line of `json.NewEncoder`), and an
API surface to learn. None of that meets the "stdlib genuinely
insufficient" bar.

Test for a proposed new dep:

1. Could the stdlib do it in a small helper file? If yes, write the
   helper.
2. Is the dep pure-Go (byob-release.8)? If no, hard stop.
3. Does the dep itself follow a similar stdlib-first discipline?
   Transitive bloat compounds.
4. Can you name the existing decision that justifies it — or do you
   need to file a new decision capturing the rationale?

If steps 1–4 all clear, add the dep and the corresponding decision
bead in the same PR. A new dep without a decision is the shape that
accretes into an unmaintainable `go.sum` over time.
