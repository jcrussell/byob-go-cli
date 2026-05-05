---
key: test-cleanup
---

Use `t.Cleanup(func())` instead of bare `defer` for test teardown. It
runs LIFO even on panic, composes cleanly with `t.Run` subtests, and
lives next to the resource it's tearing down rather than at the top
of the test. Cleanups registered inside a helper stay attached to the
test even after the helper returns — `defer` can't do that. Subtests
inherit `t` and can register their own cleanups independently.
