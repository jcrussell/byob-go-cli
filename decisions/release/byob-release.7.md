---
id: byob-release.7
title: 'Recommended .golangci.yml: pin the lint floor byob idioms assume'
type: decision
priority: 2
status: open
parent: byob-release
labels:
  - release
---

## Description

Problem: byob-release.3 names `golangci-lint run` as the body of `make lint`,
but golangci-lint v2 ships with most useful linters disabled by default.
Without a pinned `.golangci.yml`, the rules byob's idioms quietly assume
(consistent initialism casing, exported-doc shape, errcheck on every
return) aren't enforced — agents drift, fresh code accumulates lint
debt, and the memory layer that names these rules is paper protection.

Idea: ship a recommended `.golangci.yml` that enables the linters byob
relies on, with a small allowlist of overrides for cases the template's
idioms generate (e.g. `noctx` off in `_test.go` because test helpers
legitimately use background context).

Enable, at minimum:
- `errcheck` — every returned error is handled or annotated.
- `staticcheck` — SA bug-checks plus the ST naming rules, especially
  ST1003 (initialism casing).
- `revive` with `exported`, `var-naming`, `receiver-naming`,
  `unused-receiver` — covers the doc-comment-shape and the Get-prefix /
  initialism naming memories.
- `gocritic` — opinionated correctness/style.
- `govet` with `nilness` and `shadow` — backstop for the typed-nil
  return-signature decision (byob-errors.5).
- `ineffassign`, `unused` — dead-code hygiene.

Plus `goimports` (formatter) for import group order matching byob's
stdlib / third-party / project convention.

Tradeoffs: more friction on legacy code being ported into a byob-shaped
layout. byob targets greenfield Go CLI tools, so the cost is small at
template-instantiation time and only grows if you defer adoption. The
config is opt-in (the file ships in the template; targets edit it).

When not to use: never. If you remove the lint floor you also remove
the `golangci-lint` invocation from `make lint`; otherwise the makefile
target lies about what it's checking.

## Design

```yaml
# .golangci.yml — pinned defaults for byob-shaped Go CLIs.
# Schema is golangci-lint v2.

version: "2"

linters:
  default: none
  enable:
    - errcheck
    - staticcheck
    - revive
    - gocritic
    - govet
    - ineffassign
    - unused

  settings:
    staticcheck:
      checks:
        - all
        - "-ST1000"  # package comment — not mandated on internal pkgs
    revive:
      rules:
        - name: exported
        - name: var-naming
        - name: receiver-naming
        - name: unused-receiver
    govet:
      enable:
        - nilness
        - shadow

  exclusions:
    rules:
      # Test helpers legitimately use context.Background() and unkeyed
      # struct literals for fakes — silence those classes in _test.go.
      - path: _test\.go
        linters:
          - noctx
          - govet

formatters:
  enable:
    - goimports
  settings:
    goimports:
      local-prefixes:
        - github.com/<your-org>/<bin>
```

Replace `<your-org>/<bin>` at template-instantiation time so the
project-local import group lands in the third position
(stdlib / third-party / project).

