---
key: principles-dry-kiss-solid
---

Duplicate twice before abstracting. Three similar lines is better than
a premature interface or helper — the right abstraction is usually
visible only after the second or third instance lands. The byob
decisions already encode the abstractions worth pulling forward
(Factory DI, narrow Options, small interfaces, semantic error types),
so reach for them when the shape matches, not because a named
principle (KISS, YAGNI, DRY, SOLID) demands it. Treat those names as
tiebreakers during refactor, not constraints during design — invoking
them up front tends to push toward more structure rather than less.
