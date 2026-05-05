---
key: defer-unlock
---

After acquiring a mutex, defer the release on the very next line:
`mu.Lock(); defer mu.Unlock()`. Same rule for `RUnlock`, channel
closes, file `Close`, and any "must release" pair — adjacent placement
makes the unlock visible at the lock site and impossible to forget on
early returns or panics. The `tx.Rollback` exception: defer the
rollback then call `tx.Commit()` before return; rollback after a
successful commit is a documented no-op in `database/sql`, so the
defer is safe.
