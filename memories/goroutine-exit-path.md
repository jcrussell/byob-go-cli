---
key: goroutine-exit-path
---

Every `go func()` has a documented exit path: ctx cancellation via
`select { case <-ctx.Done(): return }`, a `sync.WaitGroup` the parent
calls Wait on, a channel close that breaks the inner loop, or natural
completion of a bounded loop. No fire-and-forget. A goroutine blocked
on an unbuffered send or receive is invisible to the GC and lives
forever. byob's default fanout is `errgroup.Group` (byob-lifecycle.3);
the rule applies to any goroutine.
