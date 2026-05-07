---
id: byob-storage.2
title: 'Two-field DB config: Driver + DSN, not a fused URL'
type: decision
priority: 2
status: open
parent: byob-storage
labels:
  - config
  - storage
---

## Description

Problem: a single fused URL like `DATABASE_URL=sqlite:///path.db` or
`postgres://user:pass@host/db` seems tidy but re-invents URL parsing
in user code. sqlite file paths have hairy edge cases (`:memory:`,
`file:...?mode=memory&cache=shared`, relative vs absolute). Postgres
DSNs can be URL-form or libpq keyword-form. MySQL uses its own
`user:pass@tcp(...)` format. Scheme-dispatch code drifts into a full
parser as users hit edge cases.

Idea: match stdlib `sql.Open(driver, dsn)` directly. Config exposes
two fields — `Driver string` and `DSN string` — populated from env
vars (`DB_DRIVER`, `DB_DSN`) or TOML keys. The storage factory reads
them, calls `sql.Open`, and dispatches to the per-backend constructor.
Each driver parses its own DSN in whatever format it prefers; byob
code never touches string-level parsing.

Tradeoffs: users set two env vars / TOML keys instead of one URL. In
return: no scheme bikeshedding, no re-parsing edge cases, and the
factory code is five lines. If a user really wants URL convenience
later, a tiny `ParseURL(s) (driver, dsn string, error)` helper can be
added on top without changing the core config shape.

## Design

```go
// internal/config/config.go
type DB struct {
    Driver string `toml:"driver"` // "sqlite" | "pgx"
    DSN    string `toml:"dsn"`    // driver-specific
}

// internal/storage/open.go
type Store interface { /* ... consumer-narrow methods ... */ }

func Open(ctx context.Context, cfg config.DB) (Store, error) {
    switch cfg.Driver {
    case "sqlite":
        return sqlite.Open(ctx, cfg.DSN)
    case "pgx":
        return postgres.Open(ctx, cfg.DSN)
    default:
        return nil, fmt.Errorf("unknown db driver: %q", cfg.Driver)
    }
}
```

Sample DSNs:

```
# sqlite on disk
DB_DRIVER=sqlite
DB_DSN=file:state.db?_pragma=journal_mode(WAL)

# postgres
DB_DRIVER=pgx
DB_DSN=postgres://user:pass@host:5432/appdb?sslmode=require

# cockroachdb (postgres wire, different port + params)
DB_DRIVER=pgx
DB_DSN=postgres://user:pass@cockroach.host:26257/appdb?sslmode=require
```

Factory exposes the `Store` via a lazy `sync.OnceValues` closure (see
byob-factory-di.1) so `mytool --help` doesn't pay the cost of opening a
connection.

