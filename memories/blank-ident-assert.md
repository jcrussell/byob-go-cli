---
key: blank-ident-assert
---

Declare `var _ Iface = (*Concrete)(nil)` at package scope wherever a
concrete type implements an interface you care about. The assignment
is type-checked at compile time, so any drift (renamed method, wrong
signature, new interface method) fails the build at that line instead
of at a distant runtime call site. Zero runtime cost.
