---
id: byob-output.3
title: Generate reference docs (Markdown, man pages) from the cobra tree
type: decision
priority: 2
status: open
parent: byob-output
labels:
  - cli
  - go
  - output
---

## Description

Problem: hand-written man pages and Markdown docs drift out of sync the
moment you add a flag or rename a subcommand. Users read stale examples
and file bugs. Maintaining a docs site that mirrors your CLI is a
second job.

Idea: the `github.com/spf13/cobra/doc` package walks your cobra tree and
emits documentation in several formats: Markdown (`GenMarkdownTree`), man
pages (`GenManTree`), reStructuredText (`GenRestTree`), and YAML
(`GenYamlTree`). Point it at your root command and it produces one
document per subcommand, with Usage, Flags, Inherited Flags, Examples,
and See-Also cross-links. Hook it up as a `go generate` target or a
CI step so the docs site is rebuilt on every change.

Your help text becomes the authoritative source: the `Short`, `Long`, and
`Example` fields you fill in on each command are the doc body. Writing
them well is now doubly rewarded — better `--help` output *and* better
published docs.

Tradeoffs: you lose free-form prose structure in the docs — everything is
generated from command help. For most tools that's a feature, not a bug.
If you need a narrative tutorial alongside the reference, write that as a
separate hand-maintained page and keep the generated reference
authoritative.

When not to use: single-command tools where `--help` is already enough.
Also skip if you've deliberately chosen a docs style (storybook, walkthrough)
that doesn't map to one-page-per-command.

## Design

```go
// tools/gendocs/main.go — invoked via `go run ./tools/gendocs ./docs/reference`
package main

import (
    "os"
    "github.com/spf13/cobra/doc"
    pkgcmd "mytool/pkg/cmd/root"
    "mytool/pkg/cmdutil/factory"
)

func main() {
    root := pkgcmd.NewCmdRoot(factory.New())
    root.DisableAutoGenTag = true // skip "Auto generated..." footer

    out := os.Args[1]
    if err := os.MkdirAll(out, 0o755); err != nil { panic(err) }
    if err := doc.GenMarkdownTree(root, out); err != nil { panic(err) }
}
```

```go
//go:generate go run ./tools/gendocs ./docs/reference
package root
```

For man pages:

```go
header := &doc.GenManHeader{
    Title: "MYTOOL", Section: "1",
}
_ = doc.GenManTree(root, header, "./docs/man")
```

