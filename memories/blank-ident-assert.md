---
key: blank-ident-assert
---

Declare `var _ Iface = (*Concrete)(nil)` at package scope wherever a
concrete type implements an interface you care about. The underscore
discards the value but the assignment is type-checked: if `*Concrete`
stops satisfying `Iface` (renamed method, wrong signature, new method
added to the interface), the package fails to build at that line — not
at some distant runtime call site. Cheap, zero runtime cost, catches
the "forgot to update the impl after changing the interface" bug class
the moment it's introduced.
