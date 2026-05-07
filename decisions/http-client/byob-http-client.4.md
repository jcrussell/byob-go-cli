---
id: byob-http-client.4
title: Test HTTP via httptest.NewServer, not gock/httpmock
type: decision
priority: 2
status: open
parent: byob-http-client
labels:
  - http
  - testing
---

## Description

Problem: HTTP mocking libraries (`gock`, `httpmock`) monkey-patch
the default transport or install package-level interceptors. That
fights byob-testing.1 (inject test doubles through the Factory, never
monkey-patch globals) and hides real behavior — connection reuse,
header handling, content-length mismatches — behind a mock layer
that doesn't match production.

Idea: tests construct a real server with `httptest.NewServer`,
point the client at its `.URL`, and write a `http.Handler` that
returns the canned responses. The Factory's HTTPClient closure is
overridden with one whose base URL is the test server. This mirrors
byob-storage.6 (real sqlite backend in tests) — same discipline, same
payoff: the test exercises the production code path.

Tradeoffs: slightly more code per test than `gock`'s
`Mock().Get().Reply()`. In exchange, tests catch real issues
(response-body-not-closed, incorrect Content-Type parsing, retry
behavior under server 503) that a mock library abstracts away. Real
network loopback on localhost is fast enough that parallel tests
don't care.

## Design

```go
func TestListItems(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/items" {
            http.Error(w, "not found", http.StatusNotFound); return
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintln(w, `{"items":[{"id":1,"name":"a"}]}`)
    }))
    t.Cleanup(srv.Close)

    f := testFactory(t, srv.URL) // Factory whose HTTPClient points at srv.URL
    items, err := listItems(t.Context(), f)
    // ... assertions
}
```

For flaky-network simulations (retry tests), respond 503 with a
`Retry-After` on the first attempt and 200 on the second, counting
requests with an atomic.

