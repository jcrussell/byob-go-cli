# Credits and Lineage

The decisions and memories in this template trace back to a handful of
upstream sources. Everything is rewritten generically so the template
drops into any Go CLI project without dragging upstream conventions or
code along. Credit for the underlying ideas belongs to the authors
below; mistakes and over-generalizations in the distillation belong to
this repository.

## github.com/cli/cli (the `gh` CLI)

Most of the architectural patterns in this template originate from the
`gh` CLI codebase. The `gh` team did the hard thinking for:

- Central `Factory` struct with lazy-closure dependencies injected into
  every command
- The `Options` + `NewCmdXxx(f, runF)` + pure runFunc three-part command
  shape, including the `runF` test-injection hook inside `RunE`
- Semantic error types (`FlagError`, `SilentError`, `CancelError`) mapped
  to distinct exit codes by a top-level runner
- `IOStreams` abstraction wrapping In/Out/ErrOut with TTY detection and
  a `NO_COLOR`-aware `ColorScheme`
- Opt-in structured export via `--json` / `--jq` / `--template` as a
  first-class output mode
- Cobra command groups for readable `--help` organization
- `ErrHint` wrapper for attaching user-facing remediation text to errors

Upstream: <https://github.com/cli/cli>

## spf13/cobra

Most of this template uses `cobra` as its command substrate. Several
idioms in the cobra codebase reward explicit surfacing rather than
leaving agents to discover them the hard way:

- Ship shell completions via cobra's auto-generated `completion <shell>`
  subcommand
- Set `SilenceUsage` and `SilenceErrors` on the root to stop cobra from
  dumping the usage blob on runtime errors
- `PersistentPreRunE` on the root command as app-wide middleware (auth,
  config load, logging init)
- Generate reference docs (Markdown, man pages) from the cobra tree via
  the `cobra/doc` package
- `MarkFlagsMutuallyExclusive` / `MarkFlagsRequiredTogether` /
  `MarkFlagsOneRequired` as the declarative way to validate flag
  relationships (integrates with shell completion)

Upstream: <https://github.com/spf13/cobra>

## Go source tree (standard library + `cmd/go`)

A second set of idioms comes directly from the Go source. The Go project
is one of the most idiomatic Go codebases in existence, and several
patterns are stdlib-endorsed or demonstrated by `cmd/go` itself:

- `signal.NotifyContext` for graceful Ctrl-C handling via context
  cancellation
- `context.Context` threaded through every runFunc
- `t.Helper()`, `t.Cleanup()`, `t.TempDir()` for test ergonomics
- `fmt.Errorf("...: %w", err)` as the canonical error-wrap verb
- `fs.FS` + `fstest.MapFS` as a testable filesystem seam
- `flag.Value` / `pflag.Value` for structured custom flag types
- `sync.OnceValue[T]` / `sync.OnceValues[A, B]` for lazy, type-safe
  memoization

Upstream: <https://github.com/golang/go>

## Effective Go

<https://go.dev/doc/effective_go> is somewhat dated — some of its advice
(package-level `init()` functions, certain concurrency idioms) has aged
out of modern practice — but three stated conventions still match
current practice and are worth codifying:

- Accept interfaces, return concrete types
- Compile-time interface assertions with the blank identifier
  (`var _ Iface = (*Concrete)(nil)`)
- Error messages: lowercase, no trailing punctuation, no newlines, so
  they compose cleanly under wrapping

## Go Code Review Comments wiki

<https://go.dev/wiki/CodeReviewComments> is the community-maintained
list of style rules Go reviewers cite. Most of the always-on style
memories in this template (`receiver-name`, `no-get-prefix`,
`doc-comment-shape`, `no-blank-error-discard`, `goroutine-exit-path`,
`context-first-param`, `pass-by-value-default`, `got-want-order`,
`errors-message-style`, `initialism-casing`) are distillations of rules
stated there. The wiki is also the authoritative source for the
"avoid in-band error values" guidance behind `byob-errors.5`.

## Google Go Style

<https://google.github.io/styleguide/go/decisions> is the public
Google Go style guide, structured as a set of "decisions" with
rationale (the same shape as the `decisions/` tree in this template).
It restates and extends the Code Review Comments wiki; most memories
that cite the wiki cite this guide too. It is the source for the
`quote-strings-in-errors` rule and contributes to `byob-errors.5` and
`byob-testing.2`.

## Third-party libraries and tools

A few external libraries/tools are named in decisions where the
template picks a default implementation. Each `See the …` link goes
to the decision epic where the choice (and its swap-out story) is
spelled out in full.

- `charmbracelet/huh` — prompter backend. See the `prompter` epic.
  Upstream: <https://github.com/charmbracelet/huh>
- `charmbracelet/bubbles` — unknown-total progress spinner (via
  `bubbles/spinner` + `bubbletea`). See the `progress` epic.
  Upstream: <https://github.com/charmbracelet/bubbles>
- `schollz/progressbar/v3` — known-total progress bar. See the
  `progress` epic. Upstream: <https://github.com/schollz/progressbar>
- `goreleaser/goreleaser` — release pipeline (cross-compile matrix,
  archives, checksums). See the `release` epic. Upstream:
  <https://github.com/goreleaser/goreleaser>
