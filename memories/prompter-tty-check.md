---
key: prompter-tty-check
---

Every Prompter method must check `IO.IsStdinTTY()` first. If stdin
is not a terminal (CI, piped input, nohup'd background process),
return the sentinel `ErrNotTTY` instead of prompting. Callers surface
a clear error ("pass --yes to run non-interactively") or branch on
`errors.Is(err, prompt.ErrNotTTY)`. Failure mode to avoid: a prompt
library reads EOF, interprets it as "no," and silently skips a
destructive action. Loud failure beats silent wrong answer.
