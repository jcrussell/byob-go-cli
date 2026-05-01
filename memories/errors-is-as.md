---
key: errors-is-as
---

Compare wrapped errors with `errors.Is(err, sentinel)` for sentinel
checks and `errors.As(err, &target)` to extract a typed error — never
`==` and never raw type assertion when the error might be wrapped.
Once any layer wraps with `%w` (per `errors-wrap-w`), `err == ErrFoo`
returns false even though `errors.Is(err, ErrFoo)` returns true. The
type-assertion form `e, ok := err.(*MyErr)` has the same failure
mode: a `fmt.Errorf("...: %w", &MyErr{...})` makes the assertion fail
despite the typed value being in the chain. `errors.As` walks the
chain and unwraps until it finds a match — and pairs with `errors.Is`
as the only safe pattern once `%w`-wrapping enters the codebase.
