---
id: byob-runtime-directories.3
title: 'Atomic writes: temp-in-same-dir + fsync + rename; flock for read-modify-write'
type: decision
priority: 2
status: open
parent: byob-runtime-directories
labels:
  - state
---

## Description

Problem: two failure modes bite concurrent CLI invocations:

1. **Torn writes.** `os.WriteFile(path, data, perm)` can crash
   mid-write and leave a truncated file. A reader in another
   invocation sees half a JSON document.
2. **Lost updates.** Two processes doing read-modify-write on the
   same state file can each read the old version, each write their
   update, and silently clobber each other — both "atomic renames,"
   one lost edit.

Idea: two disciplines, applied where each matters.

- **Atomic write (all state files):** create a temp file in the
  **same directory** as the target (cross-dir renames aren't atomic
  on POSIX), write the full payload, `f.Sync()`, then
  `os.Rename(tmp, final)`. Also `fsync` the parent directory after
  rename for durability on kernel crash. Wrap in a `WriteFileAtomic`
  helper so every call site gets it right.
- **File lock (read-modify-write):** advisory lock via `flock(2)`
  (use `golang.org/x/sys/unix.Flock` on POSIX;
  `syscall.LockFileEx` on Windows) held for the duration of the
  read-compute-write cycle. Lock file is a sibling of the state file
  (`state.json.lock`). Alternative: a CAS-via-version field in the
  state itself (retry on conflict) — useful when you can't take a
  lock (e.g. state on a network filesystem).

Tradeoffs: flock is advisory — well-behaved processes respect it;
rogue ones ignore it. That's fine for a single CLI talking to its
own state. For true multi-writer scenarios, sqlite (byob-storage) gives
you real transactions. JSON state + flock is the "small, structured
state" answer; sqlite is the "more than ~1KB of structured state"
answer.

## Design

```go
// pkg/cmd/paths/atomic.go
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil { return err }
    tmpName := tmp.Name()
    defer os.Remove(tmpName) // no-op on success after rename

    if _, err := tmp.Write(data); err != nil {
        tmp.Close(); return err
    }
    if err := tmp.Sync(); err != nil {
        tmp.Close(); return err
    }
    if err := tmp.Close(); err != nil { return err }
    if err := os.Chmod(tmpName, perm); err != nil { return err }
    if err := os.Rename(tmpName, path); err != nil { return err }

    // fsync parent dir for durability. POSIX-only — on Windows a
    // directory-handle Sync() isn't the same durability barrier and
    // some filesystems return errors for it. Matches the scope in
    // the atomic-rename-samedir memory.
    if runtime.GOOS != "windows" {
        d, err := os.Open(dir)
        if err != nil { return err }
        defer d.Close()
        if err := d.Sync(); err != nil { return err }
    }
    return nil
}

// pkg/cmd/paths/lock_unix.go
//go:build unix

func WithLock(path string, fn func() error) error {
    f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o644)
    if err != nil { return err }
    defer f.Close()
    if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil { return err }
    // Explicit unlock. f.Close() would also release the flock on
    // exit; keeping LOCK_UN makes the release point visible and
    // survives a refactor that defers close elsewhere.
    defer unix.Flock(int(f.Fd()), unix.LOCK_UN)
    return fn()
}

// pkg/cmd/paths/lock_windows.go
//go:build windows
//
// Use LockFileEx from golang.org/x/sys/windows with
// LOCKFILE_EXCLUSIVE_LOCK; unlock via UnlockFileEx. Same contract
// as WithLock above.
```

Pairs with the `atomic-rename-samedir` memory — one-line rule,
always-on.

