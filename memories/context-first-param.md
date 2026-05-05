---
key: context-first-param
---

`ctx context.Context` is always the first parameter of any function
that takes one, conventionally named `ctx`, never stored as a struct
field. Storing ctx on a struct ties every method's cancellation
lifetime to whoever populated the field, defeats
`errgroup.WithContext`'s derived-ctx pattern, and tempts closures to
capture a stale ctx. The byob-lifecycle.1 decision threads ctx
through every runFunc — this is the underlying rule.
