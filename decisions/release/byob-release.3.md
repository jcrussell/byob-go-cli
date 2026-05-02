---
id: byob-release.3
title: Makefile is the day-to-day entrypoint
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - cli
  - go
  - release
---

## Description

Problem: `go build` + `go test` + `go install` works for single-file
demos. Real tools want `make build`, `make test`, `make lint`,
`make clean` plus a way to invoke the release pipeline locally
(dry-run an archive build). A shell script works; a Makefile works
better for target dependencies and partial rebuilds.

Idea: the Makefile owns every dev entrypoint. Targets:

- `build` — host binary at `./bin/mytool` with ldflags (byob-release.1)
  and reproducibility flags (byob-release.6).
- `install` — `go install` with the same flags.
- `test` — `go test -race ./...`.
- `lint` — `golangci-lint run` (when configured).
- `clean` — remove `bin/` and cached artifacts.
- `release` — shell out to `goreleaser release --clean` (byob-release.4).
- `snapshot` — `goreleaser build --snapshot --clean` for local
  dry-run of the cross-compile matrix.

Version / commit / date are computed once at the top of the Makefile
from `git describe --tags --always --dirty` and exported as
`LDFLAGS`. The Makefile injects that string into dev builds;
goreleaser (byob-release.4) computes its own `{{.Version}}` from the
release tag and injects that into release builds. The two strings
are intentionally different: `make build` on a post-tag or dirty
tree produces `v1.0.0-5-g1234567-dirty`, while goreleaser on the
same tag produces `1.0.0`. The invariant worth preserving is not
"same string" but "every binary stamps its own version via ldflags
on `internal/<bin>cmd/build`" — so `mytool version` always reports
the build's own provenance regardless of which path built it.

Tradeoffs: Makefile syntax is finicky (tab indentation, recursive
expansion, `.PHONY` discipline). The payoff is a muscle-memory
entrypoint every Go developer already uses. Alternative: a
`task`/`just`/`mage` file — each less standard than Makefile.

## Design

```makefile
BIN        := mytool
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     := $(shell git rev-parse HEAD 2>/dev/null || echo none)
DATE       := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PKG        := github.com/acme/$(BIN)/internal/$(BIN)cmd/build

# Dev LDFLAGS keep symbols so delve / pprof / stack traces work.
# byob-release.6 handles the -s -w strip in .goreleaser.yml for release.
LDFLAGS    := -X $(PKG).Version=$(VERSION) \
              -X $(PKG).Commit=$(COMMIT) \
              -X $(PKG).Date=$(DATE)
# Not named GOFLAGS — that is a reserved Go env var (the toolchain
# prepends its value to every `go` invocation). Exporting it would
# double-apply -trimpath.
GO_BUILD_FLAGS := -trimpath
export LDFLAGS VERSION COMMIT DATE

.PHONY: build install test lint clean release snapshot
build:
	CGO_ENABLED=0 go build $(GO_BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BIN) ./cmd/$(BIN)

install:
	CGO_ENABLED=0 go install $(GO_BUILD_FLAGS) -ldflags "$(LDFLAGS)" ./cmd/$(BIN)

test:
	go test -race ./...

release:
	goreleaser release --clean

snapshot:
	goreleaser build --snapshot --clean
```

Cross-compile matrix does **not** live in the Makefile — that's
goreleaser's job (byob-release.4). The Makefile orchestrates; goreleaser
does the matrix.

