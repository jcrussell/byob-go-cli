---
id: byob-input-validation.2
title: Validate parsed config shape explicitly before trusting it
type: byob
priority: 2
status: open
parent: byob-input-validation
labels:
  - input-validation
---

## Description

Problem: `yaml.Unmarshal([]byte(data), &cfg)` succeeds with the zero
value for any missing field and silently ignores unknown fields by
default. Code downstream reaches into `cfg.HTTP.Timeout` expecting a
positive duration and gets 0, which HTTP clients interpret as "no
timeout." The bug surfaces hours later in production.

Idea: after parsing, every config struct has a `Validate() error`
method that runs before the Factory hands the config out. Two
discipline points make this work:

- **Reject unknown fields at decode time.** `yaml.NewDecoder(r).KnownFields(true)`
  or `json.NewDecoder(r).DisallowUnknownFields()` — typos in the
  config fail loudly instead of being silently ignored.
- **Zero is a valid "unset" only where intended.** Required fields
  use pointer types (`Timeout *time.Duration`) so a missing value is
  distinguishable from an explicit zero. `Validate()` dereferences
  and range-checks.

Prefer hand-rolled `Validate()` methods over struct-tag libraries
(`go-playground/validator`). The tags move validation out of Go and
into string literals; the methods keep logic co-located with the
struct, show up in godoc, and let you compose validation naturally
via substruct calls.

Tradeoffs: a few lines per config struct. Pays off the first time a
deployment fails cleanly at startup instead of producing corrupted
state in hour three.

When not to use: config structs with one scalar field. The ceremony
doesn't earn its keep until there are 3+ fields.

## Design

```go
// internal/config/http.go
type HTTP struct {
    Endpoint string         `yaml:"endpoint"`
    Timeout  *time.Duration `yaml:"timeout"`    // pointer: "unset" vs "0" distinguishable
    Retries  int            `yaml:"retries"`
}

func (h *HTTP) Validate() error {
    if h.Endpoint == "" {
        return errors.New("http.endpoint is required")
    }
    if _, err := url.Parse(h.Endpoint); err != nil {
        return fmt.Errorf("http.endpoint: %w", err)
    }
    if h.Timeout == nil {
        return errors.New("http.timeout is required")
    }
    if *h.Timeout <= 0 {
        return errors.New("http.timeout must be positive")
    }
    if h.Retries < 0 || h.Retries > 10 {
        return fmt.Errorf("http.retries out of range: %d (0-10)", h.Retries)
    }
    return nil
}

type Config struct {
    HTTP HTTP `yaml:"http"`
    // ...
}

func (c *Config) Validate() error {
    if err := c.HTTP.Validate(); err != nil {
        return fmt.Errorf("http: %w", err)
    }
    // ...
    return nil
}

// internal/config/load.go
func Load(r io.Reader) (*Config, error) {
    dec := yaml.NewDecoder(r)
    dec.KnownFields(true) // reject unknown keys
    var c Config
    if err := dec.Decode(&c); err != nil {
        return nil, fmt.Errorf("parsing config: %w", err)
    }
    if err := c.Validate(); err != nil {
        return nil, fmt.Errorf("config: %w", err)
    }
    return &c, nil
}
```

