---
id: byob-storage.1
title: Per-backend Store implementations; no query builder, no ORM
type: decision
priority: 2
status: open
parent: byob-storage
labels:
- cli
- go
- interfaces
- storage
---

## Description

Problem: a CLI that wants both sqlite (zero-setup default) and
postgres/cockroachdb (shared-backend deployments) has to decide where
dialect differences live. ORMs hide SQL and break when you need a
precise query. Query builders drag in a DSL to re-learn. Shared SQL
with placeholder rebinding drifts silently as dialect-specific syntax
(upsert, JSON operators, RETURNING) creeps in and only one side gets
updated.

Idea: one consumer-scoped `Store` interface per feature (see
byob-interfaces.1), with per-backend implementations under
`internal/storage/sqlite/` and `internal/storage/postgres/`. Each
backend owns its driver import, its DSN pragmas and pool sizing, its
SQL text, its placeholder style, its upsert syntax. The `Store`
interface is the only shared surface. Factory picks the backend at
startup (see the `{Driver, DSN}` config decision).

Multi-tenant variant: if the CLI is multi-tenant (one binary serving
multiple users/projects/cities), consider a two-tier `RootStore` +
`RootStore.ForTenant(id) Store` pattern so the tenant ID doesn't
thread through every method. Skip it if you're single-tenant; most
CLIs are.

When to graduate: if the schema crosses ~10 tables, the same method
diverges across backends twice, or a second contributor starts editing
SQL — that's the signal to reach for sqlc (sqlc.dev) for codegen across
per-dialect .sql files. Until then, hand-written is smaller, more
readable, and has no generated files in git.

Tradeoffs: two implementations of every write method means ~2x
write-path code. Contract tests are mandatory to catch drift. The
payoff: no DSL, no codegen step, every query visible at its call site.

## Design

```go
// pkg/cmd/list/list.go — interface lives in the consumer package
type listStore interface {
    ListItems(ctx context.Context) ([]Item, error)
}

// internal/storage/sqlite/store.go
type Store struct{ db *sql.DB }

func Open(ctx context.Context, dsn string) (*Store, error) {
    db, err := sql.Open("sqlite", dsn+
        "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)")
    if err != nil { return nil, err }
    db.SetMaxOpenConns(1) // sqlite serializes writes anyway
    return &Store{db: db}, migrate(ctx, db)
}

func (s *Store) ListItems(ctx context.Context) ([]Item, error) {
    rows, err := s.db.QueryContext(ctx,
        `SELECT id, name FROM items ORDER BY id`)
    // ... sqlite-flavored scanning
}

// internal/storage/postgres/store.go
type Store struct{ db *sql.DB }

func Open(ctx context.Context, dsn string) (*Store, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil { return nil, err }
    // driver defaults for pool sizing; tune if needed
    return &Store{db: db}, migrate(ctx, db)
}

func (s *Store) ListItems(ctx context.Context) ([]Item, error) {
    rows, err := s.db.QueryContext(ctx,
        `SELECT id, name FROM items ORDER BY id`)
    // ... postgres-flavored scanning; $1,$2,... for methods that take args
}
```

Pragmas and pool sizing stay inside each `Open` — they never leak
into shared config. The production sqlite and postgres Stores both
satisfy the consumer's narrow `listStore` interface structurally,
without importing the consumer package.

