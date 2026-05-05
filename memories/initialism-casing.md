---
key: initialism-casing
---

Initialisms (URL, HTTP, ID, JSON, API) keep consistent case in
identifiers: `appID` not `appId`, `serveHTTP` not `serveHttp`,
`URLParser` not `UrlParser`. Matches the stdlib (`http.ServeHTTP`,
`url.URL`). The trap is copy-pasting from JS/Java/C# corpora where
`userId` and `urlParser` read as natural — Go's stdlib has trained
readers to expect the all-or-nothing form, and inconsistency stands
out.
