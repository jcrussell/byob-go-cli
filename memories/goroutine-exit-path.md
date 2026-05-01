---
key: goroutine-exit-path
---

Every `go func()` you spawn has a documented exit path: ctx cancellation
via `select { case <-ctx.Done(): return }`, a `sync.WaitGroup` the
parent calls Wait on, a channel close that breaks the inner loop, or
natural completion of a bounded loop. No fire-and-forget. The Code
Review Comments wiki and the Google Go Style decisions both call out
goroutine leaks as the most common Go bug class: a goroutine blocked on
an unbuffered send or receive is invisible to the GC and lives forever.
byob's default fanout primitive is `errgroup.Group` (byob-w71.3) because
it bundles ctx cancellation, error collection, and bounded concurrency
in one library — but the underlying rule applies to any goroutine.
