---
id: byob-agent-onboarding
title: Agent onboarding
type: byob
priority: 2
status: open
labels:
  - meta
  - onboarding
---

## Description

You are working inside a forked copy of **byob-go-cli**, a personal
template repository for Go CLI tools. The beads database you're looking
at contains the template's architectural decisions alongside the tasks
for this specific project.

## Design

### What's in the workspace

- **byob library beads** (`bd list --type=byob`) — ~125 records
  shipped by the template under a custom `byob` type, organized
  into category roots (parentless) with child decisions. These are
  the preferred structure and idioms. They are NOT work to be
  closed. Treat them as design references to consult when
  implementing tasks.
- **Memories** (`bd memories`) — one-line tips (like "use
  sync.OnceValue for lazy singletons", "wrap errors with %w"). These
  auto-inject into your session via `bd prime`, so you should already
  see them in your context.
- **Task beads** (`bd ready --exclude-type=byob`, `bd list --type=task`)
  — the actual work for this specific tool. These are what you
  claim, implement, and close.

### How to work

1. At the start of each session, `bd prime` gives you the current
   workflow context and any memories.
2. Run `bd ready --exclude-type=byob` to find tasks ready to work
   on. Claim one with `bd update <id> --claim`. (Without the
   filter, byob's ~125 library beads flood the ready list; your
   own decision/epic/task beads are unaffected.)
3. Before implementing, consult the relevant byob beads. To find
   them:
     - `bd list --type=byob --no-parent` to see the category roots
     - `bd list --type=byob -l errors` (or any other category
       label) to filter by topic
     - `bd show <id>` for the full Problem / Idea / Tradeoffs / Sketch
4. Implement the task following the decisions + memories. Code should
   match the template's idioms.
5. When done, `bd close <task-id>` and move to the next ready task.
6. If you hit a gap in the template — an architectural question that
   the existing decisions don't answer — you can add a new decision
   bead for the specific tool, or (if it's a generally-reusable
   insight) surface it so the template itself can be updated.

### Key commands

- `bd list --type=byob --no-parent` — browse the category roots
- `bd list --type=byob -l errors` — drill into one category
- `bd memories` — list the tip layer
- `bd memories error` — search memories by keyword
- `bd show <id>` — full bead contents
- `bd ready --exclude-type=byob` — tasks ready to work on
- `bd create "<title>" -t task` — add a new task
- `bd prime` — re-inject workflow context (also runs automatically on
  session start)

### Philosophy

- byob beads are preferences, not contracts. They're the template's
  default answers to "how should this be structured?" — apply them
  in new code you write. Don't build anything — tests, lints, CI
  gates, pre-commit hooks, runtime asserts, custom vet checkers —
  that *fails* when a byob decision is violated. The gap between
  idiom and invariant is the whole point of the preference framing.
  Existing code that diverges might or might not be a bug; assess
  case-by-case rather than reflexively migrating.
- The template is opinionated about cobra, pure-Go (`CGO_ENABLED=0`),
  and the gh-CLI-lineage factory/command/IO idioms. Don't fight those.
- Tasks are what's unique about THIS project. byob beads are what's
  shared across all projects built from this template.

Read `CREDITS.md` for the upstream lineage (gh CLI, Go stdlib, Effective
Go, cobra) if you want the provenance of the ideas in this library.

