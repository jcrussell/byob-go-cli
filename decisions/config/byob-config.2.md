---
id: byob-config.2
title: 'Layered config with provenance: env > file > default'
type: decision
priority: 2
status: open
parent: byob-config
labels:
- cli
- config
- go
---

## Description

Problem: `mytool config get foo` returns `bar` — from where? The env? The
file? The default? Without provenance, users can't debug config issues.

Idea: resolve values through a documented layer order:
(1) environment variable override (`MYTOOL_FOO`)
(2) project/host config file
(3) built-in default

Each resolved value carries its source ("env:MYTOOL_FOO", "file:/path.toml",
"default"). A `--show-source` flag on `config get` prints the source.

Tradeoffs: slightly more machinery than a flat struct. Alternative: track
only values, lose debuggability. Rule: if you support config files AND env
overrides AND defaults, you owe users provenance.

## Design

```go
type Value[T any] struct {
    V      T
    Source string // "env:MYTOOL_FOO" | "file" | "default"
}

func ResolveString(envKey, fileVal, def string) Value[string] {
    if v := os.Getenv(envKey); v != "" {
        return Value[string]{v, "env:" + envKey}
    }
    if fileVal != "" {
        return Value[string]{fileVal, "file"}
    }
    return Value[string]{def, "default"}
}

// usage:
endpoint := ResolveString("MYTOOL_ENDPOINT", fileCfg.Endpoint, "https://api.example.com")
fmt.Printf("%s (from %s)\n", endpoint.V, endpoint.Source)
```

