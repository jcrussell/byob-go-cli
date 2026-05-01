---
key: doc-comment-shape
---

Doc comments on exported declarations start with the identifier name
and form a complete sentence: `// Store persists items to a backing
SQLite file.` not `// This struct holds the store.` The Code Review
Comments wiki and the Google Go Style decisions both mandate this shape
because godoc and IDE hover tooltips render the comment verbatim — a
comment that doesn't name what it documents loses its anchor when read
out of context. `revive`'s `exported` rule catches the violation if the
lint floor (byob-aws.7) is in place. Same rule for package comments:
`// Package store persists items to disk.`
