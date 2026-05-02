---
id: byob-runtime-directories.4
title: 'State schema versioning: refuse newer, upgrade older'
type: decision
priority: 2
status: open
parent: byob-runtime-directories
labels:
  - cli
  - go
  - state
---

## Description

Problem: a tool at v2 reading a state file written by v3 (a user
downgraded, or ran two versions on the same machine) can silently
drop fields it doesn't understand. A tool at v2 reading a v1 file
can fail to parse because a required field is missing. Both failure
modes are silent by default; both corrupt the user's state.

Idea: every persistent state file (JSON, TOML, whatever) carries a
top-level `schema_version` integer field. Writers stamp the current
version. Readers:

- **`stored > current`**: refuse to parse. Print an error naming
  the stored version, the current version, and suggest updating the
  tool. Do NOT touch the file.
- **`stored < current`**: upgrade-on-read. A per-version migration
  function transforms the file from `v1 → v2 → v3 → current`.
  Writer-side write-back happens on the next normal save; readers
  don't need to mutate on disk.
- **`stored == current`**: parse normally.

SQLite state (byob-storage) handles schema versioning via its own
migrations; this decision is specifically for JSON/TOML files that
sit outside a DB.

Tradeoffs: every schema change requires a migration function. That's
the discipline that prevents silent data loss. For state that's
purely a cache (regenerable), skip the versioning — just ignore
unreadable files and regenerate. Versioning is for state you can't
recompute.

## Design

```go
type State struct {
    SchemaVersion int    `json:"schema_version"`
    LastRepo      string `json:"last_repo,omitempty"`
    // ...
}

const CurrentSchemaVersion = 3

func Load(path string) (*State, error) {
    data, err := os.ReadFile(path)
    if err != nil { return nil, err }

    // Peek at the version first.
    var probe struct{ SchemaVersion int `json:"schema_version"` }
    if err := json.Unmarshal(data, &probe); err != nil {
        return nil, fmt.Errorf("parsing state version: %w", err)
    }
    if probe.SchemaVersion > CurrentSchemaVersion {
        return nil, fmt.Errorf(
            "state file version %d is newer than tool version %d; upgrade the tool",
            probe.SchemaVersion, CurrentSchemaVersion)
    }

    for v := probe.SchemaVersion; v < CurrentSchemaVersion; v++ {
        data, err = migrations[v](data)
        if err != nil { return nil, fmt.Errorf("migrating state v%d→v%d: %w", v, v+1, err) }
    }

    var s State
    if err := json.Unmarshal(data, &s); err != nil { return nil, err }
    // In-memory SchemaVersion reflects the migration target, not what's
    // on disk. The file still holds probe.SchemaVersion until the next
    // Save rewrites it with CurrentSchemaVersion.
    s.SchemaVersion = CurrentSchemaVersion
    return &s, nil
}

var migrations = map[int]func([]byte) ([]byte, error){
    1: migrateV1toV2,
    2: migrateV2toV3,
}
```

Writers always write `CurrentSchemaVersion`. Next save reconciles
the on-disk version with what's in memory.

