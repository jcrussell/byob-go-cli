---
key: context-first-param
---

`ctx context.Context` is always the first parameter of any function
that takes one, conventionally named `ctx`, never stored as a struct
field. Stated in the Code Review Comments wiki and the Google Go Style
decisions; the convention is universal enough that linters, reviewers,
and tooling all assume it. Storing ctx on a struct ties the
cancellation lifetime of every method to whoever populated the field,
defeats `errgroup.WithContext`'s derived-ctx pattern, and tempts
closures to capture a stale ctx. The byob-w71.1 decision threads ctx
through every runFunc — this memory is the underlying rule that
decision relies on.
