---
id: byob-factory-di.1
title: Central Factory with lazy closures for expensive dependencies
type: byob
priority: 2
status: open
parent: byob-factory-di
labels:
  - factory-di
---

## Description

Problem: package-level globals (config, DB handle, HTTP client) make
commands hard to test, hide the dependency surface, and force lifecycle
decisions at init time. At the other extreme, eagerly constructing every
dependency in `main()` means `mytool --version` and `mytool --help` pay
the cost of opening a database they never touch.

Idea: define a `Factory` struct that holds every cross-cutting
dependency. Cheap dependencies (IOStreams, prompter) are eager fields.
Expensive dependencies (config, store, HTTP client) are
`func() (T, error)` closures — lazily invoked only by commands that
actually need them. The factory is constructed once in `main()` and
threaded into every command constructor; no command ever touches a
global.

This gives you three wins in one shape:
1. **Testability** — swap factory fields for fakes in tests, no globals
   to reassign.
2. **Explicit dependency surface** — `grep NewCmdXxx` shows every
   command's signature and what it touches via `f`.
3. **Cold-start latency** — `mytool --help` never opens the store
   because nothing calls `f.Store()` on the help path.

Tradeoffs: callers must remember `f.Store()` (invoke) rather than
`f.Store` (field access). One line of boilerplate per command constructor.
Worth it after the second command.

When not to use: a single-command tool with one dependency. At that
scale the factory is ceremony; a plain struct literal in `main()` is
fine.

## Design

```go
type Factory struct {
    IOStreams *iostreams.IOStreams          // eager, cheap
    Prompter  prompt.Prompter                // eager, cheap, interface

    Config     func() (*config.Config, error)  // lazy
    Store      func() (store.Store, error)      // lazy
    HTTPClient func() (*http.Client, error)     // lazy
}

func New() *Factory {
    ios := iostreams.System()
    return &Factory{
        IOStreams: ios,
        Prompter:  prompt.NewLive(ios),
        Config:    lazyConfig(),
        Store:     lazyStore(),
        HTTPClient: lazyHTTPClient(),
    }
}

// Every command takes *Factory:
func NewCmdList(f *Factory, runF func(*Options) error) *cobra.Command {
    opts := &Options{IO: f.IOStreams, Store: f.Store}
    // ...
}

// Usage inside a command:
cfg, err := f.Config()
if err != nil { return err }
s, err := f.Store()
```

See also: `sync-oncevalue` memory for the idiomatic way to implement
each lazy closure.

