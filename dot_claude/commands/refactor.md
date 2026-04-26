---
description: Refactor the code at the given path, symbol, or target description.
argument-hint: [path|symbol]
---

Refactor the code pointed to by `$ARGUMENTS`.

Requirements:
1. Use `$ARGUMENTS` as the refactor target: one or more paths, a symbol, a subsystem, or a short instruction.
2. Inspect the target and nearby callers or tests before editing.
3. Preserve behavior while improving clarity and maintainability.
4. Prefer smaller extractions, better naming, deduplication, and simpler control flow over broad rewrites.
5. Update affected tests or call sites when needed.
6. Run focused validation after the change.
7. Summarize the refactor and the validation result.
