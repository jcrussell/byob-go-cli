---
key: test-tempdir
---

Use `t.TempDir()` for any test that touches the filesystem: it creates a
per-test isolated directory and auto-cleans after the test ends. Write
fixtures inline with `os.WriteFile` rather than maintaining a committed
`testdata/` tree — tests stay hermetic, parallel-safe, and you can read
the fixture and the assertion on the same screen. Reserve `testdata/`
for truly static reference files like golden outputs or sample binary
inputs.
