---
id: byob-prompter.5
title: --yes / -y destructive-action override
type: decision
priority: 2
status: open
parent: byob-prompter
labels:
- cli
- command-shape
- go
- prompter
---

## Description

Problem: destructive actions (delete, purge, overwrite) need
confirmation when run interactively and must be bypassable in CI
scripts. Tools that invent their own flag (`--force`, `--no-confirm`,
`--non-interactive`) end up with three inconsistent flags across
subcommands.

Idea: one flag, `--yes` (short `-y`), on any command that does a
destructive action. When set, `Confirm()` returns true without
prompting. Documented convention across the tool — every destroyer
reads the same flag.

`--yes` is NOT a persistent root flag, because not every command is
destructive. It's a per-command flag, but the flag name and
semantics are template-mandated so users see a uniform surface.

Tradeoffs: users who want a single "assume yes to everything" knob
have to learn that it's per-command. That's correct — a persistent
`--yes` would silently confirm destructive side-effects in
subcommands the user didn't intend to flag.

Relationship to `ErrNotTTY` (byob-prompter.3): `--yes` short-circuits the
TTY check. A non-interactive run with `--yes` is the supported,
correct flow. A non-interactive run without `--yes` returns
`ErrNotTTY` and fails with a clear message.

## Design

```go
type DeleteOptions struct {
    // ...
    Prompter prompt.Prompter
    Yes      bool
}

func NewCmdDelete(f *Factory, runF func(context.Context, *DeleteOptions) error) *cobra.Command {
    opts := &DeleteOptions{Prompter: f.Prompter}
    cmd := &cobra.Command{
        Use:   "delete <name>",
        Short: "Delete an item (prompts unless --yes)",
        RunE: func(c *cobra.Command, args []string) error {
            if runF != nil { return runF(c.Context(), opts) }
            return deleteRun(c.Context(), opts)
        },
    }
    cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "skip confirmation prompt")
    return cmd
}

func deleteRun(ctx context.Context, opts *DeleteOptions) error {
    if !opts.Yes {
        ok, err := opts.Prompter.Confirm(ctx, "Delete permanently?", false)
        if errors.Is(err, prompt.ErrNotTTY) {
            return fmt.Errorf("refusing to delete without --yes in non-interactive mode")
        }
        if err != nil { return err }
        if !ok { return ErrCancel }
    }
    // ...perform delete
}
```

