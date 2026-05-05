---
key: errors-is-as
---

Compare wrapped errors with `errors.Is(err, sentinel)` for sentinel
checks and `errors.As(err, &target)` to extract a typed error — never
`==`, never raw type assertion. Once any layer wraps with `%w` (per
`errors-wrap-w`), `err == ErrFoo` returns false even when
`errors.Is(err, ErrFoo)` returns true; the type-assertion form
`e, ok := err.(*MyErr)` has the same failure mode. `errors.As` walks
the chain — pair with `errors.Is` once `%w`-wrapping is in the
codebase.
