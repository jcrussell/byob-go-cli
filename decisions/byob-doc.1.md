---
id: byob-doc.1
title: README is orient-and-quickstart, not a reference
type: decision
priority: 2
status: open
parent: byob-doc
labels:
- cli
- go
- user-docs
---

## Description

Problem: READMEs that list every flag, every subcommand, and every
return code duplicate information the tool already exposes via
`--help` (and that byob-8u7.3 regenerates as Markdown/man pages on
every release). The duplicated copy drifts, ages badly, and makes
the README long enough that first-time readers bounce before reaching
the interesting part.

Idea: the README has one job — get a new reader from "never heard of
this tool" to "ran it successfully once" — and then hands off to
discoverable sources for everything else. That means four sections,
in this order:

1. **What it is.** One or two sentences. No feature list.
2. **Install.** One command. Link to alternative install methods if
   they exist; don't reproduce them inline.
3. **Quickstart.** One end-to-end worked example that produces
   visible output. The reader copy-pastes and it works.
4. **Where to look next.** Links to `--help`, generated reference
   docs, the troubleshooting page (byob-doc.5), and the repo.

Explicitly **not** in the README: exhaustive flag lists, subcommand
catalogs, config-key reference, exit-code tables, API signatures.
Those come from `--help`, `<tool> completion`, `man <tool>`, and
godoc. If the reader needs them, they already know how to find them.

Tradeoffs: the README looks sparser than typical open-source projects
where the README is the manual. That's intentional. Less surface area
means less drift, and concentrated content means first-time readers
actually read it.

When not to use: tiny single-command tools where the README and
`--help` converge to the same content naturally. Everything else
benefits from the split.

## Design

```markdown
# mytool

Short sentence on what it does and for whom.

## Install

    brew install <user>/tap/mytool

Other install methods: [releases page](…/releases) and `go install`.

## Quickstart

    mytool widgets list
    # ID   NAME
    # 1    example

## Where to look next

- `mytool --help` for the full flag and subcommand reference.
- [docs/reference/](docs/reference/) — generated from cobra on every
  release (byob-doc.2).
- [docs/troubleshooting.md](docs/troubleshooting.md) — failure modes
  and recovery (byob-doc.5).
```

Anything a returning user might need — config schema, exit codes, API
surface — lives behind one of those four links, not in this file.

