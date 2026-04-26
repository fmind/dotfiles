---
description: Improve or simplify the code for the given target.
argument-hint: [path|symbol]
---

Improve or simplify the code described by `$ARGUMENTS`.

Requirements:
1. Treat `$ARGUMENTS` as a target to improve or simplify.
2. Inspect the relevant files and nearby usages before editing.
3. Do not remove the target entirely. Improve or simplify it in place.
4. Preserve behavior unless the request explicitly asks for behavioral change.
5. Make the change and run the narrowest validation you can.
6. Summarize what you simplified and what you validated.
