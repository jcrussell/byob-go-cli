---
key: no-blank-error-discard
---

Don't write `_ = err` or `_, _ = io.Copy(dst, src)` to silence an
error. Either handle it (return, log, fall back), or annotate the
discard with a comment explaining why the error is provably safe to
ignore: `_ = f.Close() // best-effort close on read-only file`.
Silent discards are how production bugs hide in plain sight — every
one is either a missed return path or worth one line of justification.
