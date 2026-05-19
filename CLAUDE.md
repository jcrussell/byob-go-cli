# Agent Instructions

This is **byob-go-cli**, a template repository for Go CLI tools. It
ships a set of architectural decisions and idiomatic-tip memories as
`bd` (beads) records that can be imported into any Go project.

You are here for one of two reasons:

- **(a)** You've been asked to *apply* byob to a target repo — seed
  that repo's beads database with byob's decisions + memories. Start
  with the section below.
- **(b)** You're *editing the template itself* — adding or updating a
  decision in `decisions/*.md` or a memory in `memories/*.md`. See
  [README.md](./README.md#updating-the-template) for the
  `make import` / `make export` roundtrip.

If neither applies, you are probably in the wrong repo.

## Applying byob to an existing repo

**STOP.** Before reading any of the target repo's source, seed its
beads database with byob's decisions and memories. Run these commands
against the *target* repo, not against byob itself:

```bash
cd <target-repo>

# 1. Make sure the target has a beads database. Skip this if .beads/
#    already exists — bd import is an upsert and won't clobber
#    existing issues.
[ -d .beads ] || BD_NON_INTERACTIVE=1 bd init --prefix <project-prefix>

# 2. Register byob's custom issue type (additive — preserves any
#    types.custom you've already configured). bd import rejects the
#    library records without this.
current=$(bd config get types.custom 2>/dev/null)
case "$current" in *"(not set)"*) current="" ;; esac
bd config set types.custom "${current:+$current,}byob"

# 3. Download the latest byob release artifact and import it.
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/latest/download/issues.jsonl
bd import /tmp/byob-decisions.jsonl

# 4. Confirm the decisions and memories landed.
bd prime                                    # memories auto-inject here
bd list --type=byob --no-parent | head      # ~20 category roots
bd list --type=byob -l errors               # drill into one category
bd ready --exclude-type=byob                # your own work, byob hidden
bd show $(bd list -l onboarding --json | jq -r '.[0].id')
```

Pin to a specific template version with a tag URL instead of
`latest`:

```bash
curl -L -o /tmp/byob-decisions.jsonl \
  https://github.com/<user>/byob-go-cli/releases/download/v1.0.0/issues.jsonl
```

Once seeding succeeds, file task beads under the target's own prefix
for any concrete gaps you spot. **Each task bead should reference the
relevant byob decision** (e.g. `byob-command-shape`, `byob-errors`) so future
agents land on the rationale without re-deriving it.

### Anti-patterns

- **Do NOT** open with a fix-it punch list against the target's
  existing code. The decisions are reference material the project
  owner should have available, not a checklist to walk on their
  behalf. **Seed first, file beads second, implement third.**
- **Do NOT** copy byob's `CLAUDE.md` or `README.md` into the target.
  Those are byob-native documents. The target should have its own
  agent docs (written by its owner) that point at `bd prime` and the
  onboarding bead.
- **Do NOT** commit byob's `.beads/issues.jsonl` into the target's
  git tree. After `bd import`, the decision records live inside the
  target's beads DB; the jsonl file is just a transport format and
  can be deleted from `/tmp` afterwards.

## Editing the template

The markdown files under `decisions/` and `memories/` are the source
of truth. `.beads/issues.jsonl` is a build artifact produced by
`make export` and published as a GitHub release asset; it is **not**
tracked in git. See [README.md](./README.md#updating-the-template)
for the full `make import` / `make export` workflow and the CI drift
gate that enforces it.

After cloning byob fresh, initialize a local beads database with
`--skip-agents` so `bd init` doesn't re-inject its boilerplate into
this file:

```bash
BD_NON_INTERACTIVE=1 bd init --prefix byob --skip-agents
```

### Memory vs. decision: which tier?

Both tiers live in this template; they differ by budget and how
agents discover them.

- **Memory** (`memories/<key>.md`) — one paragraph, always-on.
  `bd prime` injects every memory in full at session start, so this
  tier has a real context budget that scales linearly with the
  corpus. Reserve it for idioms that compress to "RULE: X. WHY:
  short rationale" and apply broadly.
- **Decision** (`decisions/<slug>/<id>.md`) — Problem / Idea / Tradeoffs /
  Sketch, queried on demand via `bd list --type=byob -l errors`
  (or any category label) and `bd show <id>`. Unlimited growth
  budget — adding a new decision costs nothing until an agent
  actually looks at it.

If a candidate recipe has any of these, it's a decision, not a memory:

- tradeoffs or variants worth naming
- a "the catch is…" caveat
- multiple method names or API surfaces to choose among
- a code sketch longer than one line

When in doubt, start as a decision. Memories earn their always-on
slot by being universally applicable and compressible.

## Non-Interactive Shell Commands

**ALWAYS use non-interactive flags** with file operations to avoid
hanging on confirmation prompts.

Shell commands like `cp`, `mv`, and `rm` may be aliased to include
`-i` (interactive) mode on some systems, causing the agent to hang
indefinitely waiting for y/n input.

```bash
# Force overwrite without prompting
cp -f source dest           # NOT: cp source dest
mv -f source dest           # NOT: mv source dest
rm -f file                  # NOT: rm file

# For recursive operations
rm -rf directory            # NOT: rm -r directory
cp -rf source dest          # NOT: cp -r source dest
```

Other commands that may prompt:
- `scp` / `ssh` — use `-o BatchMode=yes`
- `apt-get` — use `-y`
- `brew` — use `HOMEBREW_NO_AUTO_UPDATE=1`
