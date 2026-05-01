---
id: byob-3e7
title: HTTP client
type: epic
priority: 2
status: open
labels:
- cli
- go
- http
---

## Description

A single `*http.Client` on the Factory, built from a composable
`http.RoundTripper` middleware chain; deterministic `httptest` seams
in tests.

