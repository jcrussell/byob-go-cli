---
id: byob-inv
title: Input validation
type: epic
priority: 2
status: open
labels:
- cli
- go
- input-validation
---

## Description

Defense against untrusted input beyond the flag-value checks already in `byob-kny.3`: path traversal, config-shape validation, shell injection, SQL injection, and enum/range checks at the Options boundary.

