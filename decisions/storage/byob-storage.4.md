---
id: byob-storage.4
title: withTx(ctx, fn) helper with explicit commit/rollback branching
type: decision
priority: 2
status: open
parent: byob-storage
labels:
  - cli
  - go
  - storage
---

## Description

Problem: every write method in a Store implementation repeats
`BeginTx` / `defer Rollback` / `Commit`. The naive helper has subtle
bugs — swallowing commit-error races, leaving a transaction open on
panic, allowing silent nested "transactions" that aren't savepoints
and silently lose atomicity.

Idea: a small `withTx(ctx, db, fn)` helper per backend with pinned
semantics, not left to each method's discretion:

- **Named return on `err`** so the deferred rollback can see the final
  outcome and decide whether to roll back.
- **Rollback only if needed.** If `fn` or `Commit` returned an error,
  rollback. If commit succeeded, the deferred rollback is skipped.
  `sql.ErrTxDone` is treated as benign (tx already finalized).
- **Panic safety.** `recover()` inside the deferred block rolls back
  and re-panics, so a panic in business logic doesn't leak an open
  transaction onto the connection.
- **No nesting.** Composing writes across methods is done by passing
  `*sql.Tx` down as a function argument, not by calling `withTx`
  recursively. Calling `withTx` from inside another `withTx` grabs a
  separate connection from the pool and silently loses atomicity; the
  convention is "don't."

Tradeoffs: ~25 lines of helper per backend, duplicated across sqlite
and postgres. Acceptable since each backend already owns its own
`*sql.DB` opening story. Callers compose transactions by threading
`*sql.Tx` explicitly; a little more verbose than "just call withTx
again" but keeps the commit boundary visible in the caller.

## Design

```go
// internal/storage/sqlite/tx.go  (postgres/tx.go is the same shape)

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) (err error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
        if err != nil {
            _ = tx.Rollback() // sql.ErrTxDone if already committed; benign.
        }
    }()

    if err = fn(tx); err != nil {
        return err
    }
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit: %w", err)
    }
    return nil
}
```

Usage within a Store method:

```go
func (s *Store) UpsertItems(ctx context.Context, items []Item) error {
    return withTx(ctx, s.db, func(tx *sql.Tx) error {
        for _, it := range items {
            if _, err := tx.ExecContext(ctx,
                `INSERT INTO items (id, name) VALUES (?, ?)
                 ON CONFLICT (id) DO UPDATE SET name = excluded.name`,
                it.ID, it.Name,
            ); err != nil {
                return err
            }
        }
        return nil
    })
}
```

Composing across Store methods — pass the `*sql.Tx` down:

```go
func (s *Store) migrateItemToFolder(ctx context.Context, tx *sql.Tx, id int64, folder string) error {
    // no withTx here — caller owns the transaction
    ...
}
```

