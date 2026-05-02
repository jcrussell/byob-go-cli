---
id: byob-storage.5
title: 'Timestamps: audit fields in DDL, business-logic fields in Go'
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

Problem: a blanket "set timestamps in Go" rule gains testability but
loses correctness — client clock skew scrambles audit-log ordering,
NTP adjustments make timestamps non-monotonic, and multi-instance
CLIs produce out-of-order `created_at` values. A blanket "set
timestamps in DDL" rule loses testability — you can't control
`created_at` in a test to verify retention or TTL logic.

Idea: split by column role.

- **Audit timestamps** (`created_at`, `updated_at`) — server-side.
  `DEFAULT CURRENT_TIMESTAMP` on sqlite, `DEFAULT now()` on postgres.
  Consistent within a transaction, immune to client clock skew, zero
  chance of forgetting to set the field on a new insert path. You
  almost never need to mock creation time.
- **Business-logic timestamps** (`expires_at`, `scheduled_for`,
  `deadline`, etc.) — Go side, with an injected `Clock` interface.
  These fields represent *decisions* rather than *events*, and tests
  almost always need to control them.

Tradeoffs: two rules instead of one; requires thinking about which
bucket a new column falls into. The heuristic: "did something
happen?" → DDL default. "Was this scheduled to happen?" → Go with
injected clock. If the column answers "when did we write this row?",
that's DDL. If it answers "when should some future action fire?",
that's Go.

## Design

```sql
-- sqlite migrations/001_init.sql
CREATE TABLE items (
    id         INTEGER PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP -- business-logic: set by app
);

-- postgres migrations/001_init.sql
CREATE TABLE items (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);
```

```go
// internal/clock/clock.go
type Clock interface{ Now() time.Time }

type realClock struct{}
func (realClock) Now() time.Time { return time.Now() }

func Real() Clock { return realClock{} }

// A test clock for business-logic timestamp tests.
type Fake struct{ T time.Time }
func (f *Fake) Now() time.Time { return f.T }
```

```go
// in the Store: business-logic timestamp (expires_at) is set in Go;
// audit timestamps (created_at, updated_at) are left to the DB default.
func (s *Store) CreateItem(ctx context.Context, name string, ttl time.Duration) error {
    expires := s.clock.Now().Add(ttl)
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO items (name, expires_at) VALUES (?, ?)`, name, expires)
    return err
}
```

`updated_at` semantics: if you want it to track modifications, either
bump it explicitly on every `UPDATE` (`SET ..., updated_at = CURRENT_TIMESTAMP`)
or add a trigger per backend. Don't try to share triggers across
dialects — sqlite and postgres syntax differ and the gain is small.

