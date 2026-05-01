---
key: no-get-prefix
---

Field accessors omit the `Get` prefix: `func (u *User) Name() string`,
not `GetName()`. Setters keep `Set` because `func SetName(s string)`
has no idiomatic alternative. Action verbs are not accessors and keep
their natural name: `s3.GetObject` is correct because it performs an
RPC, not a field read. The Google Go Style decisions name the rule;
`revive`'s `var-naming` catches the field-accessor violation if the
lint floor (byob-release.7) is in place. The mental model: `Counts()` is a
noun (the count), `GetCounts()` is a Java-flavored method call.
