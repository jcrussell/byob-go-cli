---
id: byob-http-client.2
title: HTTPClient on the Factory as a lazy closure with tuned timeouts
type: decision
priority: 2
status: open
parent: byob-http-client
labels:
- cli
- factory-di
- go
- http
---

## Description

Problem: `http.DefaultClient` has no timeouts — a hung TCP handshake
hangs forever regardless of `ctx.Done()`. A separately configured
client shared across commands avoids this, but constructing it
eagerly in `main()` costs time for commands that never make a
request.

Idea: put the client construction behind a lazy closure on the
Factory (byob-factory-di.1, byob-config.3). `f.HTTPClient() (*http.Client,
error)`; `sync.OnceValues` caches the result (two-return, so
`OnceValues`, not `OnceValue`). The client itself
wraps a `*http.Transport` with explicit timeouts — `DialContext`,
`TLSHandshakeTimeout`, `ResponseHeaderTimeout`,
`ExpectContinueTimeout`, `IdleConnTimeout` — rather than relying on
`DefaultTransport` (which has some, but not `ResponseHeaderTimeout`).
Per-request cancellation still flows through `ctx`, but the transport
timeouts catch cases where ctx alone can't (network-level hangs
during handshake).

Gzip: do **not** set `Accept-Encoding` manually on outgoing requests.
stdlib's `http.Transport.DisableCompression` defaults to false, which
means the transport auto-sets `Accept-Encoding: gzip` and
auto-decompresses response bodies — you get transparent gzip for
free. Setting the header explicitly disables the auto-decompress
path and you'll read a compressed body as if it were plain text.

Response-body size limits: callers reading untrusted response bodies
should wrap the reader with `http.MaxBytesReader(nil, resp.Body,
maxBytes)` (or `io.LimitReader` for the non-HTTP-server case) before
`io.ReadAll`. A malicious or buggy endpoint returning a multi-GB
body can OOM the client otherwise. Pick a ceiling per call site —
no single default fits both "paginated list of items" and "fetch
artifact binary".

Tradeoffs: a private transport means the client isn't equivalent to
`http.DefaultClient` — users who grab the client for ad-hoc scripts
get different timeout behavior than stdlib. That's the point, but
worth a docstring.

## Design

```go
func newHTTPClient(ua string) *http.Client {
    tp := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   10 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 30 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
        IdleConnTimeout:       90 * time.Second,
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   10,
    }
    return &http.Client{
        Transport: buildTransport(ua, tp),
        Timeout:   0, // use per-request ctx, not a client-wide timeout
    }
}

// f.HTTPClient is sync.OnceValues-wrapped so the first caller pays
// the construction cost, everyone else gets the cached client.
```

Client-wide `Timeout` stays at 0 deliberately: operations with
streaming bodies (downloads, large uploads) would otherwise be
killed mid-transfer. Use `ctx` per-request to bound total time.

