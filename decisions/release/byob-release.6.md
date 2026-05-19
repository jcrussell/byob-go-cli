---
id: byob-release.6
title: 'Reproducibility flags: -trimpath always, -s -w on release only'
type: byob
priority: 2
status: open
parent: byob-release
labels:
  - release
---

## Description

Problem: `-s -w` strips the binary's symbol table and DWARF debug
info. It's the right default for a release binary (smaller download,
slightly faster load) and the wrong default for a dev build — it
breaks `delve`, `pprof`, and useful stack traces. Applying it
uniformly trades debuggability for file size everywhere.

Idea: split the flag sets:

- **Dev builds** (`make build`, `make install`, `go run`):
  `-trimpath` only. Path-free binaries, full symbols.
- **Release builds** (`make release` via goreleaser):
  `-trimpath -ldflags "-s -w"`. Stripped + path-free.

`-trimpath` stays on both because it has no debugging downside — it
just removes local filesystem paths from the binary, which is good
hygiene regardless.

Pair both with `CGO_ENABLED=0` (byob-release.8) for pure-Go, static
binaries that cross-compile without a C toolchain.

Reproducibility caveat: `-trimpath` + `-s -w` + `CGO_ENABLED=0` is
**not** byte-reproducible by itself. The `Date` ldflag alone defeats
it, and reproducible Go builds also need `SOURCE_DATE_EPOCH` handling
and a pinned Go toolchain version. This decision doesn't aim for
byte-reproducibility; it aims for small, path-free release binaries
and fully debuggable dev binaries. Pursue full reproducibility only
if downstream distributors require it.

## Design

```makefile
# Makefile — dev build keeps symbols. Not named GOFLAGS (that's a
# reserved Go env var).
GO_BUILD_FLAGS := -trimpath
LDFLAGS := -X $(PKG).Version=$(VERSION) \
           -X $(PKG).Commit=$(COMMIT) \
           -X $(PKG).Date=$(DATE)
build:
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BIN) ./cmd/$(BIN)
```

```yaml
# .goreleaser.yml — release strips symbols:
builds:
  - env: [CGO_ENABLED=0]
    flags: [-trimpath]
    ldflags:
      - -s -w
      - -X github.com/acme/mytool/internal/mytoolcmd/build.Version={{.Version}}
      # ...etc
```

Same `-trimpath`, different `-s -w`. Dev binaries are debuggable;
release binaries are lean.

