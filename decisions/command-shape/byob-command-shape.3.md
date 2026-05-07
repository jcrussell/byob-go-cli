---
id: byob-command-shape.3
title: Cobra command groups for semantic help; root only aggregates
type: decision
priority: 2
status: open
parent: byob-command-shape
labels:
  - command-shape
---

## Description

Problem: `mytool --help` that dumps 30 commands in alphabetical order is a
readability failure. Users have to know which command they want before the
help is useful.

Idea: define a small set of `cobra.Group` entries on the root command (core,
query, admin, setup, info). Each feature sets `cmd.GroupID`. Help groups
related commands visually. The root command does nothing but import feature
packages and wire their constructors into groups.

Tradeoffs: adding a new group is a root-command edit; resist the urge to
invent a group per command. Three to six groups is the sweet spot. Avoid
nesting groups arbitrarily deep.

## Design

```go
root := &cobra.Command{Use: "mytool"}
root.AddGroup(
    &cobra.Group{ID: "core",  Title: "Core commands:"},
    &cobra.Group{ID: "admin", Title: "Admin commands:"},
    &cobra.Group{ID: "info",  Title: "Info commands:"},
)

addTo := func(c *cobra.Command, group string) {
    c.GroupID = group
    root.AddCommand(c)
}

addTo(create.NewCmdCreate(f, nil), "core")
addTo(list.NewCmdList(f, nil),     "core")
addTo(doctor.NewCmdDoctor(f, nil), "info")
```

