---
id: byob-storage.3
title: Hand-rolled go:embed migration runner with per-migration transactions
type: decision
priority: 2
status: open
parent: byob-storage
labels:
  - concurrency
  - storage
---

## Description

Problem: external migration tools (goose, golang-migrate, atlas)
solve real problems — advisory locks for concurrent runners, checksum
verification, down-migrations — but the template targets small CLIs
with stable schemas where those features are unused weight. A
hand-rolled runner in ~75 lines is auditable, ships zero deps, and
integrates with `go test` trivially.

Idea: one `migrations/` directory per backend
(`internal/storage/sqlite/migrations/*.sql`,
`internal/storage/postgres/migrations/*.sql`). Embed with `go:embed`.
Numeric prefix ordering (`001_init.sql`, `002_add_foo.sql`). A
`schema_version` table tracks applied versions. **Run each
migration's body and its version-insert in a single transaction** —
this is the one thing a naive runner gets wrong. If the body applies
but the `INSERT INTO schema_version` fails (ctx cancel, disk full,
crash mid-write), the next boot re-runs the migration against a DB
that already has the changes.

When to graduate: if the CLI runs multi-instance against shared
postgres (Kubernetes init containers, multiple systemd units), switch
to pressly/goose — its `pg_advisory_lock` prevents concurrent runners
from both applying migration 001. If the team ever edits a migration
file after it's been applied in the field, the missing-checksum gap
becomes a real bug class; switch to goose then too.

Tradeoffs: no down-migrations (intentional — they're usually
write-offs in practice). No advisory locks (single-instance only). No
checksum verification (don't edit applied migrations). Scope: a
developer's laptop DB, single-instance CLIs writing to a shared
postgres, and CI pipelines.

## Design

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

func migrate(ctx context.Context, db *sql.DB) error {
    if _, err := db.ExecContext(ctx,
        `CREATE TABLE IF NOT EXISTS schema_version (
            version INTEGER NOT NULL PRIMARY KEY
        )`,
    ); err != nil {
        return fmt.Errorf("create schema_version: %w", err)
    }
    var current int
    if err := db.QueryRowContext(ctx,
        `SELECT COALESCE(MAX(version), 0) FROM schema_version`,
    ).Scan(&current); err != nil {
        return fmt.Errorf("read schema_version: %w", err)
    }

    entries, err := fs.ReadDir(migrationsFS, "migrations")
    if err != nil { return fmt.Errorf("read migrations dir: %w", err) }
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Name() < entries[j].Name()
    })

    for _, e := range entries {
        if !strings.HasSuffix(e.Name(), ".sql") { continue }
        version, err := strconv.Atoi(strings.SplitN(e.Name(), "_", 2)[0])
        if err != nil || version <= current { continue }

        body, err := migrationsFS.ReadFile("migrations/" + e.Name())
        if err != nil { return fmt.Errorf("read %s: %w", e.Name(), err) }

        // body + version-record atomic: either both land or neither.
        if err := withTx(ctx, db, func(tx *sql.Tx) error {
            if _, err := tx.ExecContext(ctx, string(body)); err != nil {
                return fmt.Errorf("exec %s: %w", e.Name(), err)
            }
            _, err := tx.ExecContext(ctx,
                `INSERT INTO schema_version (version) VALUES (?)`, version)
            return err
        }); err != nil {
            return err
        }
    }
    return nil
}
```

Placeholder note: the `INSERT INTO schema_version` uses `?` here for
sqlite. The postgres migration runner lives in
`internal/storage/postgres/migrations.go` and uses `$1`. Each backend
owns its runner; they do not share.

