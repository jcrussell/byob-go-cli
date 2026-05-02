---
id: byob-security
title: Security
type: epic
priority: 2
status: open
labels:
  - cli
  - go
  - security
---

## Description

Supply-chain and secret-handling posture: pin by hash not tag, scan for known CVEs on every push, ship SBOMs and signatures from the release pipeline, and refuse to accept secrets as flag values.

