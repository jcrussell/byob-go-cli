---
id: byob-command-shape.7
title: Implement pflag.Value for custom flag types
type: decision
priority: 2
status: open
parent: byob-command-shape
labels:
  - command-shape
---

## Description

Problem: flags whose value needs custom parsing — enums
(`--format=json|yaml|text`), comma-separated lists, URLs, key=value
pairs — are often captured as `string` and parsed inside runFunc.
That means parse errors are reported after cobra has already accepted
the flag, every command that needs the same shape repeats the parse
logic, and `--help` output shows `string` instead of a meaningful
type name.

Idea: implement `pflag.Value` (cobra) or `flag.Value` (stdlib) on a
custom type. The interface is tiny: `String() string`, `Set(string)
error`, and for pflag `Type() string`. Register the flag with
`cmd.Flags().Var(&v, "name", "usage")`. You get parse-time
validation, a custom type name in `--help`, and a reusable type
across every command that needs the same shape. Unit tests can
exercise `Set()` directly without spinning up the command.

Tradeoffs: a handful of extra lines per custom type. Pays back
immediately on the second command that uses the same shape — and on
the first unit test.

When not to use: single-command, single-use parsing so simple it's
not worth a type (a one-off `--count` that's just an int). Use
`IntVar` and move on.

## Design

```go
type Format string

const (
    FormatJSON Format = "json"
    FormatYAML Format = "yaml"
    FormatText Format = "text"
)

func (f *Format) String() string { return string(*f) }
func (f *Format) Type() string   { return "format" }
func (f *Format) Set(s string) error {
    switch Format(s) {
    case FormatJSON, FormatYAML, FormatText:
        *f = Format(s)
        return nil
    }
    return fmt.Errorf("must be one of json|yaml|text")
}

func NewCmdGet(f *Factory, runF func(*Options) error) *cobra.Command {
    opts := &Options{Format: FormatText}
    cmd := &cobra.Command{Use: "get"}
    cmd.Flags().Var(&opts.Format, "format", "output format (json|yaml|text)")
    return cmd
}
```

