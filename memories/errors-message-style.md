---
key: errors-message-style
---

Error messages start with a lowercase letter, have no trailing
punctuation, and contain no newlines or tabs. Errors get wrapped into
larger sentences — `fmt.Errorf("loading config: %w", err)` composes
into `"loading config: reading foo.toml: no such file or directory"`,
clean without mid-sentence capitals or stray periods. Multi-line
remediation text belongs in an `ErrHint` wrapper or on ErrOut, not
inside the error value.
