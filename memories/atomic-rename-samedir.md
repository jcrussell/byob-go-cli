---
key: atomic-rename-samedir
---

For atomic file writes, the temp file MUST live in the same directory
as the target — `rename(2)` is atomic only within a single filesystem,
so a rename from `/tmp/foo` to `~/.config/foo` is a non-atomic
cross-filesystem copy. Use `os.CreateTemp(filepath.Dir(target),
".tmp-*")`, write, sync, rename. Wrap in a `WriteFileAtomic` helper.
