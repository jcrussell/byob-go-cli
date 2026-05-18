---
id: byob-release.9
title: Lint config lands first; fix lint findings as they appear
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - release
---

## Description

Problem: a `.golangci.yml` added after the codebase has grown surfaces
hundreds of pre-existing findings on the first run. The realistic
responses are all bad: silence the rule globally (the lint floor
degrades), `//nolint` past every finding (the noise hides real bugs),
or schedule a multi-day cleanup PR (it never lands). The lint config
from byob-release.7 only does its job if it lands when the codebase is
small enough that the first run produces zero findings.

Idea: in any new project, `.golangci.yml` (byob-release.7) and the
`make lint` target (byob-release.3) land in the first commit alongside
the cobra scaffold. In a brown-field adoption, they land in the first
commit after `bd import` of the byob beads — before the first task bead
is implemented. CI runs `make lint` alongside `make test` on every
push, and the build breaks on findings the same way it breaks on test
failures.

The cadence after that:

- Run `make lint` before every commit, alongside `make test` — the
  same pre-commit gate byob-testing.4 step 5 names, with both checks
  enforced.
- Treat new findings as build breakage. Either fix the code or fix
  the lint config — never `//nolint` past one without a comment naming
  the specific reason it's the right call for that line.
- When adding a new linter or tightening an existing rule, do it in
  its own commit and fix the resulting findings in the same commit.
  Linter changes that defer cleanup re-create the original problem at
  a smaller scale.

Tradeoffs: upfront friction during scaffolding — the first `make lint`
after wiring it up may surface findings the agent has to fix before the
initial commit lands. That's the design: the cost is small when the
codebase is empty and grows linearly with deferral. Late adoption is
exactly what the decision exists to prevent.

When not to use: if you've decided lint isn't worth the cost for this
project, remove the `golangci-lint` invocation from `make lint`
(byob-release.3) rather than letting the target lie about what it's
checking. The shape this decision exists to prevent is the half-measure
where the lint config exists but its findings are routinely ignored —
an opt-out is honest, a lying target is not.

## Design

Wire `make lint` into CI alongside `make test`:

```yaml
# .github/workflows/ci.yml — relevant steps.
- name: Test
  run: make test
- name: Lint
  run: make lint
```

For a brown-field adoption with an existing codebase, the only honest
sequencing is:

1. Land `.golangci.yml` and `make lint`.
2. In the *same* commit (or the immediately following one), fix or
   carefully `//nolint` every finding from the first run. Each
   `//nolint` carries a comment naming why the finding is the wrong
   call for that line. None of: "TODO clean this up later."
3. Land CI enforcement.

If step 2 produces a diff that's too large to land at once, the lint
floor is the wrong shape for the codebase as it stands — narrow it
(drop a linter, add an exclusion in `.golangci.yml`) until step 2 fits
in one PR. A change that splits into "land the config, land the
cleanup" with weeks in between is exactly the deferred-cleanup failure
mode this decision exists to prevent.

