---
id: byob-command-shape.5
title: PersistentPreRunE on root for app-wide middleware
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

Problem: cross-cutting setup — logging configuration, authentication
handshake, config load, telemetry init — needs to happen before *every*
subcommand's business logic. Copy-pasting that setup into each `RunE`
or into every `NewCmdXxx` is duplicative and drifts.

Idea: put setup in `PersistentPreRunE` on the root command. Cobra's
execution order is
`OnInitialize → root.PersistentPreRunE → cmd.PreRunE → cmd.RunE → cmd.PostRunE → root.PersistentPostRunE`,
so anything on the root's `PersistentPreRunE` runs once per invocation,
before any subcommand's logic, with the parsed flags already bound. Use
it as application-wide middleware: load config, wire auth, bind stuff
onto the Factory, set up graceful-shutdown context handlers.

Skip for help / version / completion subcommands by checking
`cmd.Name()` or a command annotation — you don't want `mytool --help`
to call your auth server.

Tradeoffs: the root's `PersistentPreRunE` runs for *every* subcommand,
including ones that might not need its setup. Keep it fast, or gate it
behind a check. Alternative: per-subcommand `PreRunE`, but then you're
back to duplication.

When not to use: tiny single-command tools where setup is obvious and
lives in `main()`. The pattern pays off at 3+ subcommands that all need
the same preamble.

## Design

```go
func NewCmdRoot(f *Factory) *cobra.Command {
    root := &cobra.Command{
        Use:          "mytool",
        SilenceUsage: true,

        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Skip for meta-commands that shouldn't trigger real setup.
            switch cmd.Name() {
            case "help", "completion", "version":
                return nil
            }

            // Load config (still lazy under the hood; first access wins).
            cfg, err := f.Config()
            if err != nil { return err }

            // Wire logging level from config.
            setupLogging(f.IOStreams.ErrOut, cfg.LogLevel)

            // Anything else every command needs: auth, telemetry, etc.
            return nil
        },
    }
    // add children...
    return root
}
```

