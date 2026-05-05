---
key: test-helper
---

Call `t.Helper()` as the first line of every test helper function.
The testing framework skips helper frames when reporting failure
locations, so a failed assertion points at the test that called the
helper instead of at the helper's internals. Skip it only when
debugging the helper itself.
