---
key: decisions-lookup
---

byob ships its library under a custom `byob` issue type — reference
material for architectural choices you make *now* (error patterns,
command shapes, config loaders, interface seams). Browse the ~20
category roots with `bd list --type=byob --no-parent`; drill in
with `bd list --type=byob -l errors` (or any other category label);
then `bd show <id>` for the full Problem / Idea / Tradeoffs /
Sketch. Use `bd ready --exclude-type=byob` to keep the library out
of the ready-work list — your own decisions and epics still
surface because they use the built-in types.

byob beads are preferences, not contracts. Apply them when writing
*new* code. Existing code that diverges might or might not be a
bug — assess case-by-case; don't reflexively migrate just because
something doesn't follow a byob idiom. File a task bead if the gap
is worth tracking. And don't build anything — tests, lints, CI
gates, hooks, runtime asserts — that *fails* when a byob decision
is violated. They're idioms for new code, not invariants to
enforce.
