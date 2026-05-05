---
key: sync-oncevalue
---

Use `sync.OnceValue[T]` (Go 1.21+) to wrap an expensive lazy factory
instead of hand-rolling `sync.Once` + captured vars. First call runs
the underlying function and caches the return; subsequent calls
return the cache. `sync.OnceValues[A, B]` handles the two-return
form, which is what factory closures usually want:
`func() (Store, error)`. Errors are sticky — a failed open won't
magically succeed on retry, almost always what you want for a
factory.
