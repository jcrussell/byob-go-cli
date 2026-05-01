---
key: initialism-casing
---

Initialisms (URL, HTTP, ID, JSON, API) keep consistent case in
identifiers: `appID` not `appId`, `serveHTTP` not `serveHttp`,
`URLParser` not `UrlParser`. The Code Review Comments wiki and the
Google Go Style decisions both name this; it matches the stdlib
(`http.ServeHTTP`, `url.URL`) and `staticcheck`'s ST1003 catches the
violation if the lint floor (byob-release.7) is in place. The trap is
copy-pasting from JS/Java/C# corpora where `userId` and `urlParser`
read as natural — Go's stdlib has trained readers to expect the
all-or-nothing form, and inconsistency stands out.
