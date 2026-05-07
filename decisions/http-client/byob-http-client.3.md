---
id: byob-http-client.3
title: 'Retry policy: bounded exponential + jitter, idempotent-by-default, Retry-After aware'
type: decision
priority: 2
status: open
parent: byob-http-client
labels:
  - concurrency
  - context
  - http
---

## Description

Problem: "retry 5xx" sounds like the whole policy but isn't. Real
servers return 429 (rate-limit) with a `Retry-After` that a naive
policy ignores, 408 (request-timeout) that deserves a retry, and
modern load balancers disconnect mid-response with
`io.ErrUnexpectedEOF` that a status-code-only check doesn't see.
Retrying non-idempotent POSTs silently duplicates writes. And
retrying *any* request with a body requires the body to be
replayable — the stdlib consumes it on the first round-trip.

Idea: bounded exponential backoff with full jitter, capped at N
attempts total (default 4 — one initial + 3 retries). Retry rules:

- **By status code:** 408, 429, 500, 502, 503, 504. NOT 501/505
  (never going to work), 425 (Too Early — semantics-specific; opt-in
  if you really want it), or other 5xx by default. Honor
  `Retry-After` on 429/503 — it overrides the backoff calc.
- **By error:** `net.Error` with `Timeout()` true, `syscall.ECONNRESET`,
  `syscall.ECONNREFUSED`, `syscall.EPIPE`, `io.ErrUnexpectedEOF`
  (mid-response disconnect).
- **By method:** GET / HEAD / PUT / DELETE / OPTIONS retry by default
  (RFC 9110 idempotent set — DELETE is idempotent despite common
  misconception). POST / PATCH retry **only** if the caller opts in
  via a context key (`retry.Allow(ctx)`).
- **By body:** only retry if `req.Body == nil` or `req.GetBody != nil`.
  `http.NewRequestWithContext` sets `GetBody` automatically for
  in-memory body types (`*bytes.Reader`, `*bytes.Buffer`,
  `*strings.Reader`). For streaming bodies (file uploads), callers
  who want retry must either buffer the body themselves or accept
  that retry is disabled.
- **Always** respect `ctx.Done()` between attempts, and drain+close
  the previous response body before the next attempt. Each retry
  runs against a fresh `r.Clone(ctx)` of the request so the outer
  middleware's pointer to the original request keeps its original
  body (the `RoundTripper` contract forbids mutating the caller's
  request).

On the terminal (no-retry) return, the caller owns `resp.Body`
and is responsible for closing it — same contract as any
`RoundTripper`.

Tradeoffs: the status/error matrix is longer than "retry 5xx" but
each entry has bitten a real CLI. POST opt-in is the main surprise —
document it in the package godoc so callers know to add
`retry.Allow(ctx)` before a safely-retriable POST (e.g. idempotency-
key requests). Silently dropping retry when `GetBody` is absent is
correct behavior; a comment on the retry middleware is the only
way a caller learns this.

## Design

```go
type retryRT struct {
    next        http.RoundTripper
    maxAttempts int           // default 4 (1 initial + 3 retries)
    base        time.Duration // default 500ms
}

func (t *retryRT) RoundTrip(r *http.Request) (*http.Response, error) {
    idempotent := methodIsIdempotent(r.Method) || retry.Allowed(r.Context())
    canReplay := r.Body == nil || r.GetBody != nil

    var resp *http.Response
    var err error
    for attempt := 0; attempt < t.maxAttempts; attempt++ {
        // Clone per attempt so we don't mutate the caller's request.
        // Clone is cheap; RoundTrip's contract forbids mutation.
        req := r.Clone(r.Context())
        if attempt > 0 && r.GetBody != nil {
            body, berr := r.GetBody()
            if berr != nil { return resp, berr }
            req.Body = body
        }
        resp, err = t.next.RoundTrip(req)
        if !shouldRetry(resp, err) || !idempotent || !canReplay {
            return resp, err
        }
        if attempt == t.maxAttempts-1 { break }

        wait := backoff(attempt, t.base)
        if ra := retryAfter(resp); ra > 0 { wait = ra }
        if resp != nil {
            io.Copy(io.Discard, resp.Body) // allow conn reuse
            resp.Body.Close()
        }
        timer := time.NewTimer(wait)
        select {
        case <-timer.C:
        case <-r.Context().Done():
            timer.Stop()
            return nil, r.Context().Err()
        }
    }
    return resp, err
}
```

`backoff(attempt, base)` returns `base * 2^attempt` with full
jitter, using `math/rand/v2` (auto-seeded, 1.22+; don't fall back to
`math/rand` — it needs explicit seeding and gives identical sequences
across invocations otherwise):

```go
import "math/rand/v2"

func backoff(attempt int, base time.Duration) time.Duration {
    exp := base << attempt
    return rand.N(exp) // generic N[time.Duration] — type-preserving
}
```

`retryAfter(resp)` parses both the seconds-integer and HTTP-date
forms of the header per RFC 7231.

