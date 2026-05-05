---
key: decisions-lookup
---

Decisions (`bd list --type decision`, `bd show <id>`) are reference
material for architectural choices you make *now* — error patterns,
command shapes, config loaders, interface seams. Browse epics with
`bd list --type epic`; drill in with
`bd list --type decision -l <label>`, then `bd show <id>` for the full
Problem / Idea / Tradeoffs / Sketch. Decisions are defaults — deviate
only with a documented reason. They are NOT a todo list: existing
code that doesn't follow a decision is a signal to follow the pattern
in new work, not a refactor on the user's behalf. File a task bead if
the gap is worth tracking.
