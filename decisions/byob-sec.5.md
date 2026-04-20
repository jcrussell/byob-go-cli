---
id: byob-sec.5
title: Pin GitHub Actions by SHA; scope permissions to minimum
type: decision
priority: 2
status: open
parent: byob-sec
labels:
- cli
- go
- security
---

## Description

Problem: `uses: actions/checkout@v4` resolves to whatever commit the
`v4` tag currently points at. A compromised action (it has happened:
tj-actions/changed-files CVE-2025-30066) can exfiltrate secrets from
every workflow pinned by tag until the tag is fixed — and a
malicious tag-move is invisible in PR review. Meanwhile, most
workflows run with `permissions: write-all` by default because
nobody narrowed them.

Idea: two complementary constraints on every workflow file.

1. **Pin by commit SHA, not tag.** `uses:
   actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1`
   — the trailing comment preserves human readability; the SHA is
   what GitHub actually resolves. Dependabot's `github-actions`
   ecosystem (byob-sec.1) knows how to bump SHAs when new versions
   land.
2. **Declare minimal `permissions:`.** The workflow-level default
   should be `permissions: {}` or a narrow set (`contents: read`),
   and jobs that need more (release publish, OIDC) opt in explicitly
   at job scope. GitHub's workflow-default is "inherit from repo,"
   which is usually write-everything.

For any cloud credentials CI needs (AWS, GCP, registries), use OIDC
federation — the workflow assumes a role via its OIDC token, no
long-lived keys stored in GitHub secrets.

Tradeoffs: SHA pinning produces uglier diffs than tag pinning. That's
what the trailing `# v4.1.1` comment is for — reviewers see the
semver intent even though the resolver uses the SHA.

## Design

```yaml
# .github/workflows/ci.yml
name: ci

on: [push, pull_request]

permissions: {}   # deny-by-default at workflow scope

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read   # needed for checkout
    steps:
      - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1
      - uses: actions/setup-go@0aaccfd150d50ccaeb58ebd88d36e91967a5f35b # v5.0.1
        with: { go-version: '1.24' }
      - run: go test ./...

  release:
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    permissions:
      contents: write  # publish release assets
      id-token: write  # OIDC for cosign (byob-sec.3)
    # ...
```

