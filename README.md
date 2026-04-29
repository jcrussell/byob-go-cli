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

1. **Decisions** — architectural decisions grouped into category
   epics. Full Problem / Idea / Tradeoffs / Sketch template.
   Consulted on demand via `bd list --type decision` and `bd show`.
2. **Memories** — one-line idiomatic tips (e.g. "wrap errors with
   `%w`", "use `sync.OnceValue`", "call `t.Helper()` in test helpers").
   These auto-inject into every agent session via `bd prime`, so they
   are always-on context without ceremony.

Decisions are grouped under category epics covering the breadth of a
Go CLI — architectural choices, CLI ergonomics, testing patterns,
observability, and the packaging and release surface. Once imported,
`bd list --type epic` enumerates the categories, and
`bd list --type decision -l <category>` drills in.

> **Agents:** if you've been asked to apply byob to an existing
> repo, see [CLAUDE.md](./CLAUDE.md#applying-byob-to-an-existing-repo)
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
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/latest/download/issues.jsonl
bd import /tmp/byob-decisions.jsonl
```

The new project now has its own prefix (`mytool-*`) for future task
beads plus every byob decision and memory pre-loaded. Browse decisions
with `bd list --type decision`; see the tip layer with `bd memories`.

Write your own `AGENTS.md` for the new project — byob does not ship
agent docs into forks. A minimal starter:

```markdown
# Agent Instructions

## First thing each session

    bd prime
    bd show $(bd list -l onboarding --json | jq -r '.[0].id')

## What lives where

- `bd list --type decision` — architectural decisions inherited from
  byob-go-cli. References, never work to close.
- `bd memories` — one-line tips that auto-inject via `bd prime`.
- `bd ready` / `bd list --type task` — your actual work items.

## Build & Test

_Fill in your project's build and test commands here._
```

### Brown-field: injecting decisions into an existing project

Same recipe, minus the `bd init` (your project already has a beads
workspace):

```bash
cd ~/repos/my-existing-tool
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/latest/download/issues.jsonl
bd import /tmp/byob-decisions.jsonl
```

Beads' import is an upsert, so your existing issues and memories are
untouched. The decision beads arrive under their stable `byob-*` IDs
so they coexist cleanly with your project's native prefix.

Pin to a specific template version with a tag URL instead of
`latest`:

```bash
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/download/v1.0.0/issues.jsonl
```

## Working in a fork

The beads database in a forked project holds three kinds of records:

- `type=decision` — architectural decisions from the template. Immortal
  references, not work to close.
- `_type=memory` — one-line tips. Auto-inject into every agent session.
- `type=task` — your actual work items for this specific tool. Claim
  with `bd update <id> --claim`, close with `bd close <id>`.

`bd ready` filters to open tasks, so it skips the decision beads by
design. `bd list --type decision -l <category>` browses the decisions
by category (`factory-di`, `testing`, `errors`, etc.). `bd memories`
lists the tip layer; `bd memories <keyword>` searches.

Agents landing in a fork should start with `bd prime` (runs the
memories into context) and then read the onboarding bead via
`bd show $(bd list -l onboarding --json | jq -r '.[0].id')`. The
fork's `AGENTS.md` should point at this as the first-session
workflow.

## Updating the template

Edits to the template live in `decisions/<id>.md` (decision beads) and
`memories/<key>.md` (memory tips). Workflow:

```bash
# Option A: edit a file directly, then push into beads
$EDITOR decisions/byob-n37.1.md
make import

# Option B: edit via bd, then re-sync the md tree
bd create "New principle" -t decision --parent <epic> \
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
