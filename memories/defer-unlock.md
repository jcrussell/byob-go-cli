---
key: defer-unlock
---

After acquiring a mutex, defer the release on the very next line:
`mu.Lock(); defer mu.Unlock()`. Same rule for `RUnlock`, channel
closes, file `Close`, and any other "must release" pair. Putting the
defer adjacent to the acquire makes the unlock visible at the lock
site and makes forgotten releases on early returns or panics
impossible. The `tx.Rollback` case is the practical exception worth
knowing: defer the rollback, then call `tx.Commit()` before the
function returns — rollback after a successful commit is a documented
no-op in `database/sql`, so the defer is safe.
