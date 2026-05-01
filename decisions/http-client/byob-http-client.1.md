---
id: byob-http-client.1
title: net/http + RoundTripper middleware chain, not retryablehttp/resty
type: decision
priority: 2
status: open
parent: byob-http-client
labels:
- cli
- go
- http
---

## Description

Problem: `hashicorp/go-retryablehttp`, `go-resty/resty`, and similar
wrappers bundle retry, logging, auth, and ergonomics into one opaque
surface. That's convenient until you need to customize one layer —
then the wrapper's surface fights you. And byob-dependencies.1's pure-Go,
stdlib-first posture would rather not take a dep it can replace with
60 lines.

Idea: compose `http.RoundTripper` middlewares over
`http.DefaultTransport`. Each concern (user-agent, retry, logging)
is a separate `RoundTripper` that wraps the next. Order matters:
outer transports see the final outcome; inner transports see per-try
state. Canonical order, outer → inner:

```
UserAgent → Retry → Logging → http.DefaultTransport
```

UserAgent outermost so the header is set once on the outgoing
request and every retry attempt inherits it. Retry in the middle so
each attempt is logged as a separate round-trip and the retry loop
sees per-attempt status/error from the inner transport. Logging
innermost (nearest the wire) so it records raw wire-level events
before retry semantics interpret them.

Auth is out of scope for this pass; when added, it slots between
UserAgent and Retry (401 triggers a refresh-then-retry).

Secrets: the Logging transport MUST redact `Authorization`,
`Cookie`, `Proxy-Authorization`, `Proxy-Authenticate`, and
`Set-Cookie` header values on **both** the request and the response
before calling `slog`. A `safeHeaders(h)` helper that returns a
shallow copy with these keys replaced by `"<redacted>"` is the
minimum; a broader allowlist is safer still. Response-side is the
easy miss — request-only redaction leaves `Set-Cookie` in the log
whenever a server mints a new session. Logging request bodies is
off by default; if opted in, wrap the body reader and redact known
secret-bearing fields (`password`, `token`, `api_key`).

Tradeoffs: you write the transports yourself. Each is ~30–50 lines.
Contract: every middleware must clone the request before mutating
and must propagate `ctx.Err()` from the inner transport. The
redaction discipline is easy to forget on a new header — put the
helper in the same file as the transport and add an allowlist
test.

## Design

```go
type userAgentRT struct {
    ua   string
    next http.RoundTripper
}

func (t *userAgentRT) RoundTrip(r *http.Request) (*http.Response, error) {
    r2 := r.Clone(r.Context())
    if r2.Header.Get("User-Agent") == "" {
        r2.Header.Set("User-Agent", t.ua)
    }
    return t.next.RoundTrip(r2)
}

var sensitiveHeaders = map[string]bool{
    "Authorization":       true,
    "Cookie":              true,
    "Proxy-Authorization": true,
    "Proxy-Authenticate":  true,
    "Set-Cookie":          true,
}

func safeHeaders(h http.Header) http.Header {
    out := make(http.Header, len(h))
    for k, v := range h {
        if sensitiveHeaders[http.CanonicalHeaderKey(k)] {
            out[k] = []string{"<redacted>"}
            continue
        }
        out[k] = v
    }
    return out
}

type loggingRT struct{ next http.RoundTripper }

func (t *loggingRT) RoundTrip(r *http.Request) (*http.Response, error) {
    start := time.Now()
    resp, err := t.next.RoundTrip(r)
    attrs := []any{
        "method", r.Method, "url", r.URL.String(),
        "reqHeaders", safeHeaders(r.Header),
        "status", statusOf(resp), "err", err,
        "dur", time.Since(start),
    }
    if resp != nil {
        attrs = append(attrs, "respHeaders", safeHeaders(resp.Header))
    }
    slog.DebugContext(r.Context(), "http", attrs...)
    return resp, err
}

func buildTransport(ua string, inner http.RoundTripper) http.RoundTripper {
    return &userAgentRT{ua: ua, next:
        &retryRT{next:
            &loggingRT{next: inner}}}
}
```

