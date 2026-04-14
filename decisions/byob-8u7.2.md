---
id: byob-8u7.2
title: Opt-in structured export via --json / --jq / --template
type: decision
priority: 2
status: open
parent: byob-8u7
labels:
- cli
- go
- output
---

## Description

Problem: scripters shouldn't have to parse your TTY output with regex. Every
serious CLI needs a structured mode.

Idea: commands that print "resources" accept three opt-in flags:
`--json <fields>` (emit JSON array with the listed fields), `--jq <expr>`
(run the JSON through a jq filter before printing), `--template <tmpl>` (run
the JSON through Go text/template). Resource types implement an
`ExportData(fields) map[string]any` method that produces a stable JSON shape
independent of the TTY rendering.

Tradeoffs: maintaining the JSON schema is a contract. Version it if needed.
Alternative: `--output=json` that emits a fixed shape — simpler but less
flexible. `--json <fields>` makes the client pick what they want.

## Design

```go
type Exporter interface {
    ExportData(fields []string) map[string]any
}

func AddJSONFlags(cmd *cobra.Command, exp *exportOptions) {
    cmd.Flags().StringSliceVar(&exp.Fields, "json", nil,
        "output JSON with the given fields")
    cmd.Flags().StringVar(&exp.JQ, "jq", "",
        "filter --json output with a jq expression")
    cmd.Flags().StringVar(&exp.Template, "template", "",
        "format --json output with a Go template")
}

// in runFunc:
if opts.JSON != nil {
    data := make([]map[string]any, 0, len(items))
    for _, it := range items {
        data = append(data, it.ExportData(opts.JSON))
    }
    return writeJSON(opts.IO.Out, data, opts.JQ, opts.Template)
}
// else fall through to table renderer
```

