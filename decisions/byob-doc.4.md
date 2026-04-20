---
id: byob-doc.4
title: Release notes from git history; no hand-maintained CHANGELOG
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

Problem: a hand-maintained `CHANGELOG.md` is an ongoing duplication
of information that already exists in git history. It drifts (commits
land without CHANGELOG entries, or CHANGELOG entries outlive the code
they described), it causes merge conflicts disproportionate to its
value, and deciding what belongs in it is a recurring review
argument that would be better spent on the code.

Idea: skip the CHANGELOG. Release notes come from git history, with
commit messages treated as the authoritative record. Two mechanisms
produce the user-visible notes:

- **goreleaser's `changelog` block** — collects commits since the
  previous tag, groups them by conventional-commit prefix or by
  labels on linked PRs, and writes the grouped result to the GitHub
  release body and the `CHANGELOG.md` shipped in the release archive.
  The file is a release-time artifact, not a committed source file.
- **GitHub's "Generate release notes"** — same outcome if goreleaser
  is not in the pipeline. Configurable via
  `.github/release.yml` (category labels, excluded users).

Both approaches require the same discipline: commit messages and PR
titles must read well as user-facing release notes. A one-line
subject that says "fix widgets list crash on empty workspace" is a
release note; "fix" is not. A linter (commitlint, or just PR
template review) enforces this at the review stage.

Tradeoffs: loses the human curation that a good CHANGELOG provides —
grouping related commits, writing narrative about breaking changes.
The counter: that curation happens anyway in the release body after
goreleaser emits its draft, as a one-time edit per release. The
sustained cost of maintaining CHANGELOG.md across every merge goes
away.

When not to use: projects with a strict regulatory or contractual
obligation to produce a particular CHANGELOG format. Template
targets typically aren't in that category.

## Design

```yaml
# .goreleaser.yaml
changelog:
  use: github             # pull from GitHub PR titles/labels
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^chore:'
      - '^test:'
  groups:
    - title: 'Breaking changes'
      regexp: '(?i)breaking'
      order: 0
    - title: 'Features'
      regexp: '^feat'
      order: 1
    - title: 'Bug fixes'
      regexp: '^fix'
      order: 2
    - title: 'Other'
      order: 999
```

Alternative without goreleaser:

```yaml
# .github/release.yml
changelog:
  categories:
    - title: Breaking changes
      labels: ['breaking-change']
    - title: Features
      labels: ['feature', 'enhancement']
    - title: Bug fixes
      labels: ['bug']
  exclude:
    labels: ['chore', 'docs-only']
```

No `CHANGELOG.md` in the repo root — if a release archive contains
one (goreleaser emits it), that's fine; it's a build artifact, not
source.

