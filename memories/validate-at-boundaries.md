---
key: validate-at-boundaries
---

Validate user input once, at the boundary — `Options.Validate()` per
`byob-input-validation.5`, or the config-parse step per
`byob-input-validation.2`. Internal callers can then trust what they
receive; don't re-check args in every helper. Defensive validation deep
in internal code rots into noise, hides real bugs by making the
validation graph hard to read, and quietly assumes the rest of the
codebase is untrusted when it isn't.
