---
id: byob-security.3
title: Emit SBOMs and cosign signatures from the release workflow
type: decision
priority: 2
status: open
parent: byob-security
labels:
  - security
---

## Description

Extends the Release epic (byob-release): the goreleaser workflow that
already produces archives, checksums, and optional package channels is
the natural home for attestation artifacts as well.

Problem: a user who downloads a release binary has no way to verify
what's inside it or whether it was actually built by your release
pipeline. "Supply-chain attestation" asks two questions the release
artifact should be able to answer: (1) what dependencies went in, and
(2) was this binary produced by the claimed build.

Idea: goreleaser has first-class support for both. Add a `sboms:`
block that invokes `syft` to emit a CycloneDX SBOM per artifact, and
a `signs:` block that invokes `cosign` in keyless mode — GitHub's
OIDC token authenticates to Sigstore, so there are no long-lived
keys to manage. Both SBOM and signature upload as release assets
alongside the binary and checksum files.

Consumers verify with `cosign verify-blob --certificate-identity
<workflow-url> --certificate-oidc-issuer https://token.actions.githubusercontent.com
...`. The workflow URL is a stable string tied to the exact workflow
file that built the binary.

Tradeoffs: adds ~30 seconds to a release. Requires the release
workflow to have `id-token: write` permission for OIDC. Zero runtime
cost for users who don't verify — but the option is there for those
who do (distro packagers, security-conscious fleets).

When not to use: projects that never cut releases. Template targets
do cut releases (that's what byob-release is about), so this applies.

## Design

```yaml
# .goreleaser.yaml
sboms:
  - artifacts: archive   # one SBOM per uploaded archive
    documents:
      - "${artifact}.sbom.cdx.json"

signs:
  - cmd: cosign
    args:
      - "sign-blob"
      - "--yes"
      - "--output-signature=${signature}"
      - "--output-certificate=${certificate}"
      - "${artifact}"
    artifacts: all
    signature: "${artifact}.sig"
    certificate: "${artifact}.pem"
```

```yaml
# .github/workflows/release.yml — permissions block the workflow needs
permissions:
  contents: write       # publish release assets
  id-token: write       # OIDC for cosign keyless
```

User-side verification:

```bash
cosign verify-blob \
  --certificate mytool_v1.2.3_linux_amd64.tar.gz.pem \
  --signature mytool_v1.2.3_linux_amd64.tar.gz.sig \
  --certificate-identity-regexp "https://github.com/<org>/<repo>/.+" \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  mytool_v1.2.3_linux_amd64.tar.gz
```

