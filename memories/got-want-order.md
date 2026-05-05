---
key: got-want-order
---

Test failure messages put `got` before `want`:
`t.Errorf("Foo(%v) = %v, want %v", in, got, want)`. The `testing`
package's diff output, IDEs, and reader habit all assume that order;
reversing to "expected vs actual" works locally but breaks tooling.
Mnemonic: "got is what you got, want is what you wanted." Note that
go-cmp inverts at the call site — `cmp.Diff(want, got)` — because
the diff renders as `-want +got`; the wrapping message still reads
got-then-want.
