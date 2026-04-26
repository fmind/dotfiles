---
description: Run the narrowest real validation for the target, fix failures, and rerun it.
argument-hint: [scope|path]
---

Run the narrowest real validation for `$ARGUMENTS`, fix failures, and rerun it.

Requirements:
1. Treat `$ARGUMENTS` as an optional path, file, subsystem, failing command, or bug hint.
2. Inspect the relevant files and repo config to determine the closest real validation command instead of assuming a generic `test` script exists.
3. Prefer the narrowest executable check for the touched area: targeted tool validation, file-scoped checks, or a focused project command.
4. If no focused check exists, explain that briefly and run the safest broader validation you can justify.
5. Fix failures in code or config until it works, then rerun the same validation.
6. Keep changes minimal and avoid unrelated edits whenever possible.
7. Summarize the validation you chose, what you fixed, and any remaining blockers.
