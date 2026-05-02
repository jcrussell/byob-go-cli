---
id: byob-command-shape.4
title: Ship shell completions via cobra's completion subcommand
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

Problem: users want tab-completion for your tool's subcommands, flags, and
flag values. Hand-writing bash / zsh / fish / powershell completion scripts
is tedious, error-prone, and drifts out of sync the moment you add a new
command.

Idea: cobra builds a `completion <shell>` subcommand for you automatically
(via `InitDefaultCompletionCmd`), which emits a shell-specific completion
script to stdout on demand. Users source it — `eval "$(mytool completion
bash)"` — and get tab-completion for every subcommand, every flag, and
anything you've marked with `ValidArgs` or `ValidArgsFunction`. No
hand-written script, no drift, and it stays correct forever as you add
commands because the script is regenerated from your current command tree
at invocation time.

Pair with `ValidArgs` (static list) or `ValidArgsFunction` (dynamic, can
hit the network or read a file) to make tab-completion domain-aware:
resource names, file paths filtered by extension, enum values.

Tradeoffs: users have to wire the `eval` into their shell rc file once.
Docs for that are standard — copy the snippet cobra's own help prints out.
After that it's invisible.

When not to use: a tool with two subcommands and four flags probably
doesn't need tab completion. Everything above that threshold benefits
enormously.

## Design

```go
// Cobra creates the `completion` subcommand automatically. To customize
// per-arg completions, set ValidArgs or ValidArgsFunction:
cmd := &cobra.Command{
    Use:       "delete <name>",
    ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) (
        []string, cobra.ShellCompDirective,
    ) {
        names := fetchNames(c.Context()) // can be dynamic
        return names, cobra.ShellCompDirectiveNoFileComp
    },
}

// Register a flag's valid values:
cmd.RegisterFlagCompletionFunc("format",
    func(c *cobra.Command, args []string, toComplete string) (
        []string, cobra.ShellCompDirective,
    ) {
        return []string{"json", "yaml", "table"},
               cobra.ShellCompDirectiveNoFileComp
    })
```

Users then enable completion in their shell rc:

```bash
# bash
eval "$(mytool completion bash)"
# or persist:
mytool completion bash > /etc/bash_completion.d/mytool

# zsh
mytool completion zsh > "${fpath[1]}/_mytool"
```

