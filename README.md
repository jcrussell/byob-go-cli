# byob-go-cli

**Bring Your Own Beads — Go CLI edition.** A forkable template
repository for Go CLI tools. Holds my preferred architectural decisions
and idiomatic tips as [beads](https://github.com/steveyegge/beads)
records, ready to be cloned into a new project (or injected into an
existing one) as the starting point for coding agents.

> **Status: experimental.** The decision set, memory tier, and
> distribution mechanics are all likely to evolve. Expect breaking
> changes to category names, bead IDs, and the `make import` /
> `make export` workflow between releases. Pin to a specific tag if
> you need stability.

## What this is

This is not a library and not a Go package — it's a **template**. You
fork it by copying it, re-init its beads database with a project-
specific prefix, and then add your own task beads on top of the
inherited decision beads. Coding agents working in the fork consult the
decision beads as "how to structure things" and implement the tasks
following those decisions.

Two layers of guidance live in the beads DB:

1. **Library beads** — architectural decisions grouped into
   categories, shipped under a custom `byob` issue type so they're
   distinguishable from your project's own work. Full Problem /
   Idea / Tradeoffs / Sketch template. Consulted on demand via
   `bd list --type=byob` and `bd show`.
2. **Memories** — one-line idiomatic tips (e.g. "wrap errors with
   `%w`", "use `sync.OnceValue`", "call `t.Helper()` in test helpers").
   These auto-inject into every agent session via `bd prime`, so they
   are always-on context without ceremony.

Library beads are grouped under category roots (parentless beads)
covering the breadth of a Go CLI — architectural choices, CLI
ergonomics, testing patterns, observability, and the packaging and
release surface. Once imported, `bd list --type=byob --no-parent`
enumerates the categories, and `bd list --type=byob -l errors` (or
any category label) drills in.

> **Agents:** if you've been asked to apply byob to an existing
> repo, see
> [CLAUDE.md](https://github.com/jcrussell/byob-go-cli/blob/main/CLAUDE.md#applying-byob-to-an-existing-repo)
> first. The workflow is "seed the target's beads DB from the
> release, then file task beads" — not "review and fix."

## Quickstart

Byob is distributed as a single release artifact:
`.beads/issues.jsonl`, built by CI from the markdown under
`decisions/` and `memories/` and attached to every GitHub release.
You never `git clone` byob into your project — you `curl` the
artifact and `bd import` it. Two variants depending on whether the
target project already has a beads workspace.

### Green-field: starting a new Go CLI project

```bash
mkdir ~/repos/mytool && cd ~/repos/mytool
git init
BD_NON_INTERACTIVE=1 bd init --prefix mytool
# Register byob's custom issue type so bd import accepts the
# library records.
current=$(bd config get types.custom 2>/dev/null)
case "$current" in *"(not set)"*) current="" ;; esac
bd config set types.custom "${current:+$current,}byob"
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/latest/download/issues.jsonl
bd import /tmp/byob-decisions.jsonl
```

The new project now has its own prefix (`mytool-*`) for future task
beads plus every byob bead and memory pre-loaded. Browse the
library with `bd list --type=byob --no-parent` for the category
roots or `bd list --type=byob -l errors` to drill into one. See
the tip layer with `bd memories`.

Write your own `AGENTS.md` for the new project — byob does not ship
agent docs into forks. A minimal starter:

```markdown
# Agent Instructions

## First thing each session

    bd prime
    bd show $(bd list -l onboarding --json | jq -r '.[0].id')

## What lives where

- `bd list --type=byob` — byob library beads (architectural
  decisions inherited from byob-go-cli). References, never work
  to close.
- `bd memories` — one-line tips that auto-inject via `bd prime`.
- `bd ready --exclude-type=byob` / `bd list --type=task` — your
  actual work items.

## Build & Test

_Fill in your project's build and test commands here._
```

### Brown-field: injecting decisions into an existing project

Same recipe, minus the `bd init` (your project already has a beads
workspace):

```bash
cd ~/repos/my-existing-tool
# Register byob's custom issue type (additive — preserves any
# types.custom you've already configured).
current=$(bd config get types.custom 2>/dev/null)
case "$current" in *"(not set)"*) current="" ;; esac
bd config set types.custom "${current:+$current,}byob"
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/latest/download/issues.jsonl
bd import /tmp/byob-decisions.jsonl
```

Beads' import is an upsert, so your existing issues and memories
are untouched. The library beads arrive under their stable
`byob-*` IDs and the custom `byob` type, so they coexist cleanly
with your project's native prefix and don't collide with your
own decisions or epics.

Pin to a specific template version with a tag URL instead of
`latest`:

```bash
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/download/v1.0.0/issues.jsonl
```

## Working in a fork

The beads database in a forked project holds three kinds of records:

- `type=byob` — byob library beads (architectural decisions from
  the template). Immortal references, not work to close.
- `_type=memory` — one-line tips. Auto-inject into every agent session.
- `type=task` — your actual work items for this specific tool. Claim
  with `bd update <id> --claim`, close with `bd close <id>`.

`bd ready --exclude-type=byob` shows ready work without the
library polluting the list — your own decisions and epics still
surface because they use the built-in types.
`bd list --type=byob -l <category>` browses the library by
category (`factory-di`, `testing`, `errors`, etc.). `bd memories`
lists the tip layer; `bd memories <keyword>` searches.

Agents landing in a fork should start with `bd prime` (runs the
memories into context) and then read the onboarding bead via
`bd show $(bd list -l onboarding --json | jq -r '.[0].id')`. The
fork's `AGENTS.md` should point at this as the first-session
workflow.

## Updating the template

Edits to the template live in `decisions/<slug>/<id>.md` (decision
beads, grouped into one subdirectory per epic) and `memories/<key>.md`
(memory tips). Workflow:

```bash
# Option A: edit a file directly, then push into beads
$EDITOR decisions/command-shape/byob-command-shape.1.md
make import

# Option B: edit via bd, then re-sync the md tree
bd create "New principle" -t byob --parent byob-errors \
  --body-file body.md --design-file design.md
make export
```

`make export` rewrites every `decisions/*.md` and `memories/*.md` from
the beads DB with stable frontmatter ordering and keeps file-based
diffs paragraph-level. It also writes a local `.beads/issues.jsonl`
as a build artifact, but that file is gitignored — it's not source,
it's what CI uploads. `make import` pushes the markdown file tree
back into the DB.

Releases are tagged with git (`git tag v1.0.0 && git push --tags`);
the release workflow in `.github/workflows/release.yml` starts from
an empty beads DB, runs `make import` + `make export` to build a
fresh `issues.jsonl` from the committed markdown, and publishes it
as a release asset. The `main` branch CI (`ci.yml`) performs the
same round-trip on every push and fails if the markdown drifts —
preventing the "forgot to commit after editing" drift class. The
markdown under `decisions/` and `memories/` is the only source of
truth this repo tracks.

## Site

The decisions and memories also render as a browsable static site at
**https://jcrussell.github.io/byob-go-cli/**. The site is built by
`.github/workflows/pages.yml` on every push to `main`; locally,
`make site` produces a copy in `_site/`.

The first time the site is enabled in a fork, the repo owner needs
to flip Settings → Pages → "Build and deployment → Source" to
`GitHub Actions`. After that, pushes to `main` deploy automatically.
Forks that rename the repo should update the `--base-url` and
`--repo-url` flags in `.github/workflows/pages.yml`.

## Lineage

Decisions here trace back to four sources: the `gh` CLI codebase, the
Go standard library and `cmd/go`, Effective Go, and the `spf13/cobra`
framework. See [`CREDITS.md`](./CREDITS.md) for the full attribution.

## Philosophy

- **Template, not library.** This repository exists to be forked. It is
  not meant to be imported as a Go module.
- **Ideas over implementations.** Each bead should be reimplementable
  from its description alone. Sketches are illustrative, not canonical.
- **Architecture vs tips, separated.** Decisions carry the full
  Problem / Idea / Tradeoffs / Sketch template; memories carry
  one-paragraph tips. `bd prime` auto-injects the memories so agents
  always have them; decisions are consulted on demand.
- **Files are the source of truth.** Markdown lives under `decisions/`
  and `memories/`; the beads DB is a local working copy, regenerable
  from the md trees at any time.
- **Opinionated about cobra and pure-Go.** If you're not on cobra or
  you're willing to link against C, several patterns won't apply
  directly. Fork another template instead.

## License

BSD 3-Clause — see [LICENSE](./LICENSE).
