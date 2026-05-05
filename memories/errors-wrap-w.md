---
key: errors-wrap-w
---

Always wrap errors with `%w`, never `%v` or `%s`:
`fmt.Errorf("reading %s: %w", path, err)`. `%w` preserves the chain
so `errors.Is`/`errors.As` can find the underlying error; `%v`
stringifies and silently severs it. Printed output is identical — the
only difference is machine-readable introspection. Use `%v` only in
log messages you're building but not returning.
