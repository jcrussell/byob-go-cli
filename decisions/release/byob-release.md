---
id: byob-release
title: Release
type: epic
priority: 2
status: open
labels:
  - release
---

## Description

Pure-Go discipline (CGO_ENABLED=0, go:embed) as the foundation; Makefile+ldflags for day-to-day builds; goreleaser for tag-triggered cross-compile, archives, checksums, optional homebrew/nfpm channels. Both paths inject the same ldflags vars so version output is path-independent.

