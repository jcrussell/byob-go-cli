---
key: receiver-name
---

Method receiver names are a 1–2 letter abbreviation of the type, used
consistently across every method on that type — `func (s *Store) ...`
everywhere, never mixing in `func (st *Store)` or generic
`this`/`self`/`me`. Short receivers read like math; consistent
receivers let you scan a type's methods without re-parsing the
binding name.
