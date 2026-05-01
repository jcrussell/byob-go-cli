---
id: byob-user-docs.2
title: Reference generated from cobra; narrative docs cover concepts only
type: decision
priority: 2
status: open
parent: byob-user-docs
labels:
- cli
- go
- user-docs
---

## Description

Problem: a tool with both hand-written reference docs and
generated-from-cobra reference docs eventually has two reference
docs, and they disagree. Users file issues about the disagreement.
Maintainers paper over it with a README note about "which is
authoritative."

Idea: exactly one source for each kind of content.

- **Reference** (flags, subcommands, exit codes, config keys): always
  generated. `byob-output.3` runs `cobra/doc` on each release to emit
  Markdown per subcommand and a `man` page; those get published as
  release assets and (if the project has a docs site) rendered as
  HTML. Reference docs are never hand-edited — the drift surface
  shrinks to the cobra source strings, which live next to the code
  they describe.
- **Narrative** (concepts, workflows, troubleshooting, "why"
  documentation): always hand-written, under `docs/`. Narrative docs
  must not restate flag defaults, subcommand syntax, or config keys
  — they link to the generated reference for those and explain how
  things compose.

The practical test: if a sentence in narrative docs would need to be
updated whenever a flag's default changes, that sentence is in the
wrong place. Move it to the generated reference (via cobra's Long
description or the Example field — see byob-user-docs.3).

Tradeoffs: narrative writers lose the freedom to state flag defaults
for emphasis. That's the point; without the constraint, the two
sources drift.

## Design

Repo layout:

```
docs/
├── concepts.md          # hand-written: why/how
├── workflows/
│   ├── first-run.md     # hand-written: end-to-end flow
│   └── ci-setup.md
├── troubleshooting.md   # hand-written: failure modes + recovery
└── reference/           # generated; .gitignored
    ├── mytool.md
    ├── mytool_auth.md
    └── mytool_widgets_list.md
```

Generation step (wired into `make docs` and the release workflow):

```go
// internal/gendocs/main.go
package main

import (
    "log"
    "github.com/spf13/cobra/doc"
    "mytool/pkg/cmd/root"
)

func main() {
    root := root.NewCmdRoot(factory.New())
    if err := doc.GenMarkdownTree(root, "docs/reference"); err != nil {
        log.Fatal(err)
    }
}
```

```make
docs:
	go run ./internal/gendocs
```

Rendering integration depends on the docs site (MkDocs, Hugo, etc.) —
the generated Markdown lives under `docs/reference/` and is included
verbatim.

