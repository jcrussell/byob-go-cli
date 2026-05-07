---
id: byob-interfaces.2
title: Accept interfaces, return concrete types
type: decision
priority: 2
status: open
parent: byob-interfaces
labels:
  - interfaces
---

## Description

Problem: returning an interface type from a constructor hides the concrete
type's useful methods from callers. Taking a concrete type as a parameter
ties your function to one implementation and blocks test doubles.

Idea: reverse those defaults. Function parameters should accept the
narrowest interface that describes what they actually use (`io.Reader`,
`fs.FS`, a locally-defined `Lister`). Constructors and factories should
return concrete types (`*Factory`, `*sqlite.Store`, `*http.Client`).
Callers that only want the interface view can assign the return value to
an interface variable themselves.

Complements the "define interfaces in the consumer package" bead: together,
they say "the consumer declares a narrow interface, and the producer gives
it a concrete value that happens to satisfy that interface."

Exception: return an interface when the concrete type is a true
implementation detail that should not leak (e.g., `crc32.NewIEEE()` returns
`hash.Hash32` because the underlying type has no useful public surface).

Tradeoffs: returning concrete types means callers can depend on every
exported method of that type — which is the goal for structs you own, and
the downside for types you'd rather keep flexible. When that downside
matters, return an interface on purpose; don't do it by reflex.

When not to use: rare; this is the default posture. Deviate only when a
concrete return type would encourage coupling to internals the caller
should not care about.

## Design

```go
// Parameters: accept narrow interfaces.
func LoadConfig(fsys fs.FS, name string) (*Config, error) { ... }
func Copy(dst io.Writer, src io.Reader) (int64, error)    { ... }

// Constructors: return concrete types.
func NewFactory() *Factory                     { ... }
func Open(path string) (*Store, error)          { ... }

// Callers assign to an interface variable if they only want the narrow view:
var r io.Reader = strings.NewReader("hello")   // strings.NewReader returns *Reader

// Justified exception: the concrete type is a true implementation detail.
func NewHash32() hash.Hash32 { return crc32.NewIEEE() }
```

