---
id: byob-http-client
title: HTTP client
type: byob
priority: 2
status: open
labels:
  - http
---

## Description

A single `*http.Client` on the Factory, built from a composable
`http.RoundTripper` middleware chain; deterministic `httptest` seams
in tests.

