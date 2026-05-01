---
id: byob-security.2
title: Run govulncheck in CI on every push
type: decision
priority: 2
status: open
parent: byob-security
labels:
- cli
- go
- security
---

## Description

Problem: CVEs in Go dependencies — and in the Go standard library
itself — land in the vulnerability database continuously. Without a
scanner wired to CI, they go unnoticed until a downstream user files
an issue or a distro packager flags the binary.

Idea: `golang.org/x/vuln/cmd/govulncheck` is the official Go
vulnerability scanner. It differs from generic SBOM scanners in one
important way: it only flags CVEs in code paths that are actually
reachable from the entry points you scan. A vulnerable symbol inside
an imported package you never call is not reported. The false-positive
rate is near zero.

Run it in CI over `./...` on every push. Treat any finding as a
build failure — by the time a CVE has a govulncheck entry, it has a
fix, and the project should upgrade or vendor a patch.

Tradeoffs: the first pass on an existing codebase can surface a
backlog. Work through it once; afterward it's noise-free. Adds ~10s
to CI.

When not to use: never. This is free coverage once the toolchain
pinning from byob-security.1 is in place.

## Design

```yaml
# .github/workflows/ci.yml — added as a job alongside the existing
# drift-check job.
  vuln:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - name: govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@v1.1.3
          govulncheck ./...
```

For a local pre-commit / Makefile target:

```make
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.1.3 ./...
```

