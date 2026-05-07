---
id: byob-agent-onboarding
title: Agent onboarding
type: decision
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

- **Decision beads** (`bd list --type decision`) — durable
  architectural decisions organized into category epics. These are
  the preferred structure and idioms. They are NOT work to be closed.
  Treat them as design references to consult when implementing tasks.
- **Memories** (`bd memories`) — one-line tips (like "use
  sync.OnceValue for lazy singletons", "wrap errors with %w"). These
  auto-inject into your session via `bd prime`, so you should already
  see them in your context.
- **Task beads** (`bd ready`, `bd list --type task`) — the actual work
  for this specific tool. These are what you claim, implement, and
  close.

### How to work

1. At the start of each session, `bd prime` gives you the current
   workflow context and any memories.
2. Run `bd ready` to find tasks ready to work on. Claim one with
   `bd update <id> --claim`.
3. Before implementing, consult the relevant architectural decisions.
   To find them:
     - `bd list --type epic` to see the categories, then
       `bd list --type decision -l <category>` to filter
     - `bd show <id>` for the full Problem / Idea / Tradeoffs / Sketch
4. Implement the task following the decisions + memories. Code should
   match the template's idioms.
5. When done, `bd close <task-id>` and move to the next ready task.
6. If you hit a gap in the template — an architectural question that
   the existing decisions don't answer — you can add a new decision
   bead for the specific tool, or (if it's a generally-reusable
   insight) surface it so the template itself can be updated.

### Key commands

- `bd list --type decision` — browse all decisions
- `bd list --type decision -l errors` — decisions in one category
- `bd list --type epic` — the architectural category epics
- `bd memories` — list the tip layer
- `bd memories error` — search memories by keyword
- `bd show <id>` — full bead contents
- `bd ready` — tasks ready to work on
- `bd create "<title>" -t task` — add a new task
- `bd prime` — re-inject workflow context (also runs automatically on
  session start)

### Philosophy

- Decisions are the template's default answers to "how should this be
  structured?" Deviate only with reason, and note why.
- The template is opinionated about cobra, pure-Go (`CGO_ENABLED=0`),
  and the gh-CLI-lineage factory/command/IO idioms. Don't fight those.
- Tasks are what's unique about THIS project. Decisions are what's
  shared across all projects built from this template.

Read `CREDITS.md` for the upstream lineage (gh CLI, Go stdlib, Effective
Go, cobra) if you want the provenance of the ideas in this library.

