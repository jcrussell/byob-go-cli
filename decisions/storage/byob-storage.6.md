---
id: byob-storage.6
title: In-memory sqlite default; same Store contract suite runs on both backends
type: byob
priority: 2
status: open
parent: byob-storage
labels:
  - concurrency
  - storage
  - testing
---

## Description

Problem: `:memory:` sqlite is a delightful default for tests — fast,
parallel-safe, no Docker. But it lies. It permits type coercions
postgres rejects. Foreign-key enforcement is off unless a pragma is
set. There's no `ENUM`, no array, no `jsonb`. String sort uses a
different default collation. A green `:memory:` run only proves your
sqlite code works; it says nothing about the postgres backend.

Idea: one `Store` contract test suite in a shared
`internal/storage/storagetest/` package. Each backend's test file
invokes the suite with a `newStore(t) storage.Store` factory. Sqlite
runs always (no setup, no env, no Docker). Postgres runs if
`TEST_POSTGRES_DSN` is set; otherwise the test skips. CI is
responsible for setting `TEST_POSTGRES_DSN` on the postgres lane so
the contract tests always run against both backends in CI. Laptop
devs get a fast sqlite loop by default.

Tradeoffs: contract tests don't catch everything — ORDER BY stability,
concurrent-write semantics, and transaction isolation level differences
can still drift. But they catch the 90% case (upsert semantics, NULL
handling, FK enforcement, basic type coercion). Unit tests that never
touch SQL (business-logic tests against a fake `Store`) continue to
live next to their consumers and don't need either backend.

Open follow-up: `testcontainers-go` for automatic postgres provisioning
in CI (so developers don't need to stand up their own DB). Tracked
separately — will land as a future `testing`-labeled decision if
adopted.

## Design

```go
// internal/storage/storagetest/contract.go
package storagetest

type StoreFactory func(t *testing.T) storage.Store

// RunContract executes the shared suite against whatever Store the
// factory produces. Each backend's *_test.go file calls this.
func RunContract(t *testing.T, newStore StoreFactory) {
    t.Helper()

    t.Run("CreateAndList", func(t *testing.T) {
        s := newStore(t)
        ctx := t.Context()
        if err := s.CreateItem(ctx, "alpha", time.Hour); err != nil {
            t.Fatal(err)
        }
        items, err := s.ListItems(ctx)
        if err != nil { t.Fatal(err) }
        if len(items) != 1 || items[0].Name != "alpha" {
            t.Fatalf("got %+v", items)
        }
    })

    t.Run("UpsertRoundtrip", func(t *testing.T) { /* ... */ })
    t.Run("TxRollbackOnError", func(t *testing.T) { /* ... */ })
    t.Run("ForeignKeyEnforcement", func(t *testing.T) { /* ... */ })
}
```

```go
// internal/storage/sqlite/store_test.go
func TestStore(t *testing.T) {
    storagetest.RunContract(t, func(t *testing.T) storage.Store {
        s, err := sqlite.Open(t.Context(), ":memory:")
        if err != nil { t.Fatal(err) }
        t.Cleanup(func() { _ = s.Close() })
        return s
    })
}
```

```go
// internal/storage/postgres/store_test.go
func TestStore(t *testing.T) {
    dsn := os.Getenv("TEST_POSTGRES_DSN")
    if dsn == "" {
        t.Skip("TEST_POSTGRES_DSN not set; skipping postgres contract tests")
    }
    storagetest.RunContract(t, func(t *testing.T) storage.Store {
        s, err := postgres.Open(t.Context(), dsn)
        if err != nil { t.Fatal(err) }
        t.Cleanup(func() {
            // truncate or drop-and-recreate per test if the DSN points at a shared DB
            _ = s.Close()
        })
        return s
    })
}
```

CI config sets `TEST_POSTGRES_DSN` on a dedicated postgres job so
contract tests fail the PR if the backends diverge. Business-logic
tests against a fake `Store` (per byob-testing.1) stay hermetic and need
neither backend.

