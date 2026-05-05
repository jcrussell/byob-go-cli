---
key: pass-by-value-default
---

Don't pass pointers to function arguments just to save a copy. Small
types — int, string, time.Time, structs of a few words, slices,
maps — go by value. Passing `*string` adds an indirection, can force
the value to escape to the heap, and signals "this function might
mutate me" that you don't mean. Reach for a pointer when the function
genuinely mutates the receiver, when the struct is large enough that
copying matters (proto messages, big configs), or when nil encodes
"absent."
