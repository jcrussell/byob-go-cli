---
id: byob-aws.4
title: goreleaser owns the release cross-compile matrix
type: decision
priority: 2
status: open
parent: byob-aws
labels:
- cli
- go
- release
---

## Description

Problem: a hand-rolled Makefile that loops over GOOS/GOARCH and
produces `.tar.gz`/`.zip` archives with SHA256 checksums, a Homebrew
formula, and optional deb/rpm packages is 200 lines of shell that
drifts. Managing these concerns by hand is exactly where release
tooling pays rent.

Idea: `.goreleaser.yml` configures:

- `builds[]`: linux/darwin/windows × amd64/arm64 (extend as the
  audience requires).
- `archives[]`: tar.gz for unix, zip for windows, consistent naming.
- `checksum`: single `SHA256SUMS` file alongside the archives.
- `brews[]`: optional Homebrew tap (update a formula repo on
  release).
- `nfpms[]`: optional deb/rpm packages for linux distros.
- Same ldflags as the Makefile (byob-aws.1) for version injection.

Invoked two ways:

- `make release` → `goreleaser release --clean` (run from the
  release GitHub Action on a `v*` tag).
- `make snapshot` → `goreleaser build --snapshot --clean` for local
  dry-run without publishing.

Tradeoffs: one more tool to learn (goreleaser's YAML schema) and one
more dep in CI (`goreleaser` must be installed in the runner).
Offset: homebrew formula management, checksum files, and archive
naming conventions are exactly the kind of bookkeeping a hand-rolled
Makefile gets wrong the third time.

## Design

```yaml
# .goreleaser.yml
version: 2

builds:
  - id: mytool
    main: ./cmd/mytool
    binary: mytool
    env: [CGO_ENABLED=0]
    flags: [-trimpath]
    ldflags:
      - -s -w
      - -X github.com/acme/mytool/internal/mytoolcmd/build.Version={{.Version}}
      - -X github.com/acme/mytool/internal/mytoolcmd/build.Commit={{.FullCommit}}
      - -X github.com/acme/mytool/internal/mytoolcmd/build.Date={{.Date}}
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]

archives:
  - format: tar.gz
    format_overrides: [{ goos: windows, format: zip }]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "SHA256SUMS"
  algorithm: sha256

brews:
  - repository: { owner: acme, name: homebrew-tap }
    directory: Formula
    homepage: https://github.com/acme/mytool
    description: A CLI tool.
    license: BSD-3-Clause
```

The Makefile and goreleaser deliberately produce different
`build.Version` strings — see byob-aws.3 for the split. What they
share is the invariant that every binary's `version` output comes
from `-X` ldflags on the `internal/<bin>cmd/build` package, not
from runtime guesswork. That is the invariant worth preserving.

