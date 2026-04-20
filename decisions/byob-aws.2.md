---
id: byob-aws.2
title: version subcommand + --version flag on root
type: decision
priority: 2
status: open
parent: byob-aws
labels:
- cli
- command-shape
- go
- release
---

## Description

Problem: users expect both `mytool --version` (a one-liner) and
`mytool version` (a richer block showing commit, build date, Go
version, OS/arch). Implementing them separately drifts. Implementing
only `--version` fails the user who pastes the version subcommand
into a bug report and wants the full block.

Idea: wire `--version` to cobra's built-in `Version` field (one-line
format string); ship a `version` subcommand that prints the richer
block. Both read from the same `build.Info()` accessor (byob-aws.1)
so they can't disagree.

Format:

```
mytool <version>                      # --version (short)

mytool <version>                      # mytool version (long)
  commit: <sha>
  built:  <iso-date>
  go:     <go-version>
  os:     <goos>/<goarch>
```

Tradeoffs: two code paths for the same data; kept trivially
consistent by routing both to one formatter function. Worth the
minor duplication for the UX win.

## Design

```go
// pkg/cmd/root/root.go
root.Version = build.Info().Version
root.SetVersionTemplate("mytool {{.Version}}\n")
root.AddCommand(cmdversion.NewCmdVersion(f))

// pkg/cmd/version/version.go
func NewCmdVersion(f *Factory) *cobra.Command {
    return &cobra.Command{
        Use:   "version",
        Short: "Print version, commit, and build info",
        RunE: func(c *cobra.Command, args []string) error {
            info := build.Info()
            fmt.Fprintf(f.IOStreams.Out, "mytool %s\n", info.Version)
            fmt.Fprintf(f.IOStreams.Out, "  commit: %s\n", info.Commit)
            fmt.Fprintf(f.IOStreams.Out, "  built:  %s\n", info.Date)
            fmt.Fprintf(f.IOStreams.Out, "  go:     %s\n", runtime.Version())
            fmt.Fprintf(f.IOStreams.Out, "  os:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
            return nil
        },
    }
}
```

`version` output goes to `Out` (not ErrOut): this is data a user
would reasonably `grep` or include in a bug report.

