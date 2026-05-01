---
key: got-want-order
---

Test failure messages put `got` before `want`, in that order:
`t.Errorf("Foo(%v) = %v, want %v", in, got, want)`. Stated in the Code
Review Comments wiki and the Google Go Style decisions, and it matters
because IDEs, grep aliases, and the `testing` package's own diff output
assume it. Reversing to "expected vs actual" works locally but breaks
tooling and confuses readers who have internalized the canonical order.
The mnemonic: "got is what you got, want is what you wanted." Note that
go-cmp inverts at the call site — `cmp.Diff(want, got)` — because the
diff is rendered as `-want +got`; the message you wrap around it still
follows got-then-want.
