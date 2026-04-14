---
key: errors-wrap-w
---

Always wrap errors with `%w`, never `%v` or `%s`: `fmt.Errorf("reading
%s: %w", path, err)`. `%w` preserves the error chain so downstream
`errors.Is(err, fs.ErrNotExist)` and `errors.As(err, &pathErr)` can find
the underlying error; `%v` stringifies and silently severs the chain.
Printed output is identical — the only difference is machine-readable
introspection. Rule: if you're adding context to an error you're
returning, use `%w`. Use `%v` only in log messages you're building but
not returning.
