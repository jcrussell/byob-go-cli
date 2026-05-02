---
id: byob-input-validation.4
title: Every SQL statement uses placeholders; never string-concatenated values
type: decision
priority: 2
status: open
parent: byob-input-validation
labels:
  - cli
  - go
  - input-validation
  - storage
---

## Description

Problem: `fmt.Sprintf("SELECT * FROM items WHERE name = '%s'", name)`
with a `name` coming from anywhere user-controlled is the canonical
SQL-injection vector. It stays a problem even in small CLIs: "name"
may start as safely-typed config but later get populated from a flag
value or a parsed file. The only robust fix is making the wrong shape
syntactically impossible to write.

Idea: `database/sql`'s placeholder API is the required form.
`?` for sqlite and mysql drivers, `$1`/`$2` for postgres. The driver
substitutes the values at protocol level — they never pass through a
SQL parser as text. That makes injection impossible by construction.

The supporting decision byob-storage.1 ("per-backend Store
implementations; no query builder, no ORM") ties in here: because
queries are hand-written SQL, the discipline is visible in every
query method. A code review rule becomes "any `fmt.Sprintf` near a
SQL string is a bug." Structural linters (`go vet -vettool`,
`semgrep`) can enforce this pattern if the team wants machine
backing.

Identifiers (table names, column names) are the one case placeholders
don't cover — drivers won't parameterize them. When they must be
dynamic, validate against an allowlist: either a fixed slice of
known-good identifiers or a strict regex (`^[a-zA-Z_][a-zA-Z0-9_]*$`).
Reject anything else.

Tradeoffs: none on the value side. The identifier-allowlist rule is
the only friction, and it applies to a small fraction of queries.

When not to use: never relax. The template's storage stance
(byob-storage) assumes placeholder-only queries throughout.

## Design

```go
// WRONG: format-string interpolation, injection vulnerability.
func getItemBad(ctx context.Context, db *sql.DB, name string) (*Item, error) {
    row := db.QueryRowContext(ctx,
        fmt.Sprintf("SELECT id, name FROM items WHERE name = '%s'", name))
    // ...
}

// RIGHT: placeholder. sqlite/mysql use `?`, postgres uses `$1`.
func getItem(ctx context.Context, db *sql.DB, name string) (*Item, error) {
    row := db.QueryRowContext(ctx,
        `SELECT id, name FROM items WHERE name = ?`, name)
    var it Item
    if err := row.Scan(&it.ID, &it.Name); err != nil {
        return nil, err
    }
    return &it, nil
}

// Dynamic column name with allowlist:
var sortableColumns = map[string]bool{
    "name": true, "created_at": true, "id": true,
}

func listItemsSorted(ctx context.Context, db *sql.DB, sortBy string) (*sql.Rows, error) {
    if !sortableColumns[sortBy] {
        return nil, cmdutil.FlagErrorf("--sort must be one of name|created_at|id")
    }
    // column name is safe here because it's been validated against the allowlist
    q := fmt.Sprintf(`SELECT id, name FROM items ORDER BY %s`, sortBy)
    return db.QueryContext(ctx, q)
}
```

