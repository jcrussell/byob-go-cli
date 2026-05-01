---
id: byob-security.4
title: Secrets from env or OS keyring only; never as flag values
type: decision
priority: 2
status: open
parent: byob-security
labels:
- cli
- go
- security
---

## Description

Problem: `mytool --token=abc123` leaks in three ways at once. The
full command line lands in shell history (`~/.bash_history`,
`~/.zsh_history`), it's visible to anyone who runs `ps` or reads
`/proc/<pid>/cmdline`, and it shows up in CI job logs verbatim.
Tokens in committed config files (`.env`, `config.yaml`) leak a fourth
way: git history.

Idea: the CLI refuses to read secrets from flag values. Secrets come
from exactly two places:

1. **OS keyring** via `zalando/go-keyring` — macOS Keychain, Windows
   Credential Manager, Linux libsecret. Written by `<tool> auth
   login`; read on demand.
2. **Environment variable** with a documented name
   (e.g. `MYTOOL_TOKEN`). Accepted so CI and server deployments can
   provide credentials without a keyring.

File-based secrets are accepted only via `--token-file=<path>`
(path, not content) — the CLI reads the file itself so the secret
doesn't round-trip through the shell. Config files store references
(`token_keyring_key: mytool-token`), never values.

If neither source yields a token, fail fast with an `ErrHint`
(byob-errors.2) pointing at `<tool> auth login`. Never fall back to a
prompt that reads the token over a TTY — prompts can't guarantee the
keystrokes aren't logged.

Tradeoffs: one more abstraction (~200 lines via zalando/go-keyring).
Eliminates an entire class of leaks.

## Design

```go
// internal/auth/token.go
type TokenSource struct {
    KeyringService string // e.g. "mytool"
    EnvVar         string // e.g. "MYTOOL_TOKEN"
}

var ErrNoToken = errors.New("no credentials configured")

func (s *TokenSource) Load(ctx context.Context) (string, error) {
    if v := os.Getenv(s.EnvVar); v != "" {
        return v, nil
    }
    v, err := keyring.Get(s.KeyringService, "default")
    if err == nil && v != "" {
        return v, nil
    }
    if errors.Is(err, keyring.ErrNotFound) || err == nil {
        return "", errhint.With(ErrNoToken,
            "run `mytool auth login` to store credentials")
    }
    return "", fmt.Errorf("reading keyring: %w", err)
}

// pkg/cmd/foo/foo.go — flag registration explicitly lacks a --token flag:
cmd.Flags().StringVar(&opts.TokenFile, "token-file", "",
    "path to a file containing the token (alternative to keyring/env)")
// no: cmd.Flags().StringVar(&opts.Token, "token", "", ...)
```

