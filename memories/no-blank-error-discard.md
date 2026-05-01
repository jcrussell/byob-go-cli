---
key: no-blank-error-discard
---

Don't write `_ = err` or `_, _ = io.Copy(dst, src)` to silence an
error. Either handle it (return, log, fall back), or annotate the
discard with a comment explaining why the error is provably safe to
ignore: `_ = f.Close() // best-effort close on read-only file`. The
Code Review Comments wiki and the Google Go Style decisions both call
this out; `errcheck` in the lint floor (byob-aws.7) catches the
unannotated form. The rule exists because silent error discards are
how production bugs hide in plain sight — every one of them is either
a missed return path or worth one line of justification.
