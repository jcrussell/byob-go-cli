---
id: byob-user-docs.5
title: Troubleshooting page keyed by ErrHint anchors
type: decision
priority: 2
status: open
parent: byob-user-docs
labels:
- cli
- go
- user-docs
---

## Description

Problem: "troubleshooting docs" are either absent (users file issues
for things the maintainer has seen five times) or present but
unfindable (the user hits an error, doesn't know the maintainer
wrote a section about it, and never navigates to the right docs
page). Links from error messages to docs are the part that makes the
docs reachable from the user's actual context.

Idea: one `docs/troubleshooting.md` with stable anchors, one anchor
per named failure mode, and an `ErrHint` (byob-errors.2) on each
semantic error type that points at the corresponding anchor. The
user sees

    Error: no credentials configured
    Hint: run `mytool auth login` to store credentials
          (see: https://mytool.dev/troubleshooting#no-credentials)

and the link resolves to the exact docs section.

The page structure is "one level of heading per failure mode," no
nesting. Each section has:

- **Symptom** — the error text and any surrounding context the user
  sees.
- **Cause** — what actually happened, in one paragraph.
- **Recovery** — the specific commands to run.
- **Prevention** — if avoidable, how.

No "General troubleshooting" section. Either a failure mode has a
dedicated entry or it doesn't belong on this page.

Anchor stability is a compatibility contract: once shipped in an
`ErrHint`, an anchor outlives the error type. Renaming is a breaking
change for users on older binaries whose links would 404.

Tradeoffs: one more place to maintain. The payoff is specific:
support load drops, because users who hit an error and want to fix
it find the fix in one click from the error message.

## Design

```markdown
# Troubleshooting

<!-- Each anchor below is referenced from an ErrHint in the source.
     Do not rename anchors; users on older releases link to them. -->

## no-credentials

**Symptom.** `Error: no credentials configured`

**Cause.** `mytool` found no token in `$MYTOOL_TOKEN` or in the OS
keyring under the `mytool` service name (see byob-security.4).

**Recovery.**

    mytool auth login

**Prevention.** In CI, set `MYTOOL_TOKEN` via federated OIDC rather
than a long-lived secret.

## workspace-not-found

**Symptom.** `Error: workspace "foo" not found`

**Cause.** The workspace name passed to `--workspace` (or set via
`MYTOOL_WORKSPACE`) does not resolve against the current account.

**Recovery.**

    mytool workspaces list

**Prevention.** Use tab completion — `mytool --workspace <TAB>`
enumerates valid names (byob-command-shape.4).
```

Paired with the source-side `ErrHint`:

```go
return errors.NewHintf(
    "no credentials configured",
    "run `mytool auth login` to store credentials",
    "https://mytool.dev/troubleshooting#no-credentials",
)
```

The anchor string (`no-credentials`) is the contract. Renaming the
heading in markdown is a breaking change for every binary that shipped
the `ErrHint` pointing at it.

