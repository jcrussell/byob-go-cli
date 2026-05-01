---
key: receiver-name
---

Method receiver names are a 1–2 letter abbreviation of the type, used
consistently across every method on that type — `func (s *Store) ...`
everywhere, never mixing in `func (st *Store)` or generic
`this`/`self`/`me`. The Code Review Comments wiki and the Google Go
Style decisions both call this out: short receivers read like math,
consistent receivers let you scan a type's methods without re-parsing
the binding name. The lint floor (byob-aws.7) catches drift via
`revive`'s `receiver-naming` rule, but state the rule explicitly here
for new code and code review.
