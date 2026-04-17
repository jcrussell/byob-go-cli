---
key: decisions-lookup
---

Decisions (`bd list --type decision`, `bd show <id>`) are reference
material for architectural choices you are making *now* — picking an
error pattern, a command shape, a config loader, an interface seam.
Browse them by category with `bd list --type decision -l <label>`
(labels: errors, command-shape, iostreams, config, testing, output,
lifecycle, layout, interfaces, factory-di, deps-philosophy). Decisions
are the template's default answers; deviate only with a reason worth
writing down. They are NOT a todo list: a decision that existing code
doesn't yet follow is not a signal to refactor on the user's behalf —
only a signal to follow the pattern for new work. File a task bead if
the gap is worth tracking; otherwise leave it alone.
