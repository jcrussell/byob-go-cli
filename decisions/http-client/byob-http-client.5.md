---
id: byob-http-client.5
title: 'User-Agent: <tool>/<version> (<os>; <arch>)'
type: decision
priority: 2
status: open
parent: byob-http-client
labels:
- cli
- go
- http
- release
---

## Description

Problem: servers use the User-Agent header to identify clients for
rate-limit buckets, deprecation warnings, and log analysis. Default
`Go-http-client/1.1` is useless for ops. Hand-setting it in every
call site is repetitive and drifts.

Idea: a single `userAgentRT` middleware (see byob-http-client.1) sets the
header once on every outbound request. The UA string is computed at
client-construction time from the `build` package vars populated by
the `release` epic (byob-release.1):

```
<tool>/<version> (<goos>; <goarch>) [commit=<short>]
```

The UA string is read via `build.Info()` (byob-release.1), which already
handles the "no ldflags" fallback using `debug.ReadBuildInfo()`. No
additional VCS lookup belongs here — the build package is the single
source of truth.

Tradeoffs: coupling the HTTP client to the build package is mild
cross-cutting, but both live in the same binary and the UA is the
natural place for the coupling to surface. Alternative: pass UA
into the factory constructor — more flexible, but every caller
duplicates the formatting.

## Design

```go
import (
    "runtime"
    "myproject/internal/mytoolcmd/build"
)

// Sentinel values for unknown build metadata. Exported by the build
// package so every consumer tests against the same string.
// build.VersionDev = "dev"; build.CommitNone = "none".

func userAgent() string {
    info := build.Info()
    ua := fmt.Sprintf("mytool/%s (%s; %s)", info.Version, runtime.GOOS, runtime.GOARCH)
    if info.Commit != "" && info.Commit != build.CommitNone {
        n := 7
        if len(info.Commit) < n { n = len(info.Commit) }
        ua += " commit=" + info.Commit[:n]
    }
    return ua
}
```

