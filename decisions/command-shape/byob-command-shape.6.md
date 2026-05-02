---
id: byob-command-shape.6
title: Cobra flag-group helpers over hand-rolled validation
type: decision
priority: 2
status: open
parent: byob-command-shape
labels:
  - cli
  - command-shape
  - go
---

## Description

Problem: validating flag combinations (mutual exclusion,
required-together, at-least-one-required) inside runFunc means
validation runs *after* side effects like opening files or logging in
have already happened, error messages are hand-written and
inconsistent across commands, and shell completion has no idea which
flags conflict — so tab-complete happily offers combinations that
will fail validation.

Idea: use cobra's declarative flag-group helpers. They run in cobra's
validation phase before RunE, emit consistent error messages, and
integrate with shell completion so conflicting flags are hidden from
tab-complete:

- `cmd.MarkFlagsMutuallyExclusive("json", "yaml", "template")` — at
  most one.
- `cmd.MarkFlagsRequiredTogether("key", "secret")` — all or none.
- `cmd.MarkFlagsOneRequired("file", "stdin", "url")` — at least one.

For *value* validation (e.g., `--port` must be 1-65535), there's no
declarative helper — a runFunc check is appropriate. Wrap the error
with `cmdutil.FlagErrorf` so the top-level runner maps it to exit
code 2 (usage error) instead of 1 (generic error).

Tradeoffs: the helpers cover the three most common
flag-relationship shapes. Rarer ones ("A and B require C, but not D")
still need runFunc validation — and that's fine; don't contort a
simple check into a helper combination.

## Design

```go
func NewCmdExport(f *Factory, runF func(*Options) error) *cobra.Command {
    opts := &Options{IO: f.IOStreams}
    cmd := &cobra.Command{
        Use: "export",
        RunE: func(c *cobra.Command, args []string) error {
            if opts.Port < 1 || opts.Port > 65535 {
                return cmdutil.FlagErrorf("--port must be 1-65535, got %d", opts.Port)
            }
            if runF != nil {
                return runF(opts)
            }
            return exportRun(opts)
        },
    }
    cmd.Flags().BoolVar(&opts.JSON, "json", false, "emit JSON")
    cmd.Flags().BoolVar(&opts.YAML, "yaml", false, "emit YAML")
    cmd.Flags().StringVar(&opts.Template, "template", "", "custom template")
    cmd.Flags().IntVar(&opts.Port, "port", 8080, "listen port")

    // declarative flag relationships — validated before RunE
    cmd.MarkFlagsMutuallyExclusive("json", "yaml", "template")
    cmd.MarkFlagsOneRequired("json", "yaml", "template")

    return cmd
}
```

