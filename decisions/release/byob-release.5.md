---
id: byob-release.5
title: 'Release workflow: tag-triggered GHA invokes goreleaser'
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - release
---

## Description

Problem: running `goreleaser release` from a developer laptop works
but mints a release from unreproducible state (laptop-specific Go
toolchain, environment variables, possibly uncommitted files). A
tag-triggered GitHub Action runs from a clean checkout with a pinned
Go version.

Idea: `.github/workflows/release.yml` triggers on `v*` tag push.
Steps:

1. `actions/checkout@v4` with `fetch-depth: 0` (goreleaser needs
   full git history for `git describe`).
2. `actions/setup-go@v5` pinned to the project's current Go version.
3. `goreleaser/goreleaser-action@v6` running `release --clean`.

The workflow needs `GITHUB_TOKEN` (automatic) for release creation
and archive upload, and `HOMEBREW_TAP_TOKEN` (manual secret) if a
homebrew tap is configured.

Any project-specific release artifacts (e.g. the byob template
itself exports `.beads/issues.jsonl`) are attached as **separate**
workflow steps that upload to the same GitHub release after
goreleaser finishes — don't try to make goreleaser manage them.

Tradeoffs: the workflow is declarative and re-runnable. The runner's
Go version is pinned in the YAML, so an old release tag rebuilt
months later still uses its original toolchain. Only caveat: a tag
accidentally pushed triggers a publish. Use a release branch
convention if that's a concern.

## Design

```yaml
# .github/workflows/release.yml
name: release
on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      # Byob-only: publish the decision/memory export as a release asset.
      # Projects that aren't byob-shaped can delete this step.
      - name: upload byob issues.jsonl
        run: make export && gh release upload ${{ github.ref_name }} .beads/issues.jsonl
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

