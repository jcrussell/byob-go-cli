---
key: atomic-rename-samedir
---

To write a file atomically, the temp file MUST live in the same
directory as the target. On POSIX, `rename(2)` is atomic only
within a single filesystem — a rename from `/tmp/foo` to
`~/.config/foo` is a cross-filesystem copy that isn't atomic.
Use `os.CreateTemp(filepath.Dir(target), ".tmp-*")`, write, sync,
rename. Wrap in a `WriteFileAtomic` helper so every call site gets
it right.
