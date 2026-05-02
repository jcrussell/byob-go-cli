---
id: byob-security.1
title: Pin Go dependencies by exact version; never `@latest` in CI
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

Problem: `go get example.com/pkg@latest`, `go install tool@main`, or a
Dockerfile that installs a tool without a version all produce
non-reproducible builds. Worse, if an attacker compromises a
package's `main` branch or publishes a new tag, the next CI run picks
up the malicious version silently.

Idea: `go.mod` already pins to exact semver, and `go.sum` records the
cryptographic hash of every downloaded module. The discipline is to
never undermine that:

- No `@latest`, `@main`, or branch names in `go get` / `go install`
  anywhere that runs repeatedly (CI, Dockerfiles, Makefile targets).
  Always an explicit version: `go install
  golang.org/x/vuln/cmd/govulncheck@v1.1.3`.
- `go mod tidy` runs in CI and fails the build if it produces a
  non-empty diff — catches hand-edited `go.mod` that skipped `tidy`.
- Upgrades go through PRs (Dependabot or hand), never ad-hoc on a
  developer laptop.

Tradeoffs: pinned versions need periodic refreshing. That's what
Dependabot / Renovate are for. The alternative — floating refs — is
trading a small maintenance cost for a large supply-chain exposure.

When not to use: never. Pinning is the default stance for any build
the project will do more than once.

## Design

```bash
# CI step that fails on un-tidied go.mod:
go mod tidy
git diff --exit-code go.mod go.sum

# tool installation (note the explicit version):
go install golang.org/x/vuln/cmd/govulncheck@v1.1.3
```

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule: { interval: weekly }
  - package-ecosystem: github-actions
    directory: /
    schedule: { interval: weekly }
```

