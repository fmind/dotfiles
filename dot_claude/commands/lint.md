---
description: Run lint checks for the target or repository, fix issues, and rerun them.
argument-hint: [scope|path]
---

Run lint checks for `$ARGUMENTS`, fix issues, and rerun the relevant checks.

Requirements:
1. Treat `$ARGUMENTS` as an optional path, file, scope, or lint error hint.
2. Inspect the repo before running commands, and prefer documented lint commands from scripts, Makefiles, mise tasks, or repo docs over inventing a different top-level command.
3. When possible, run a narrower lint command for the affected files before falling back to the repo-wide lint task.
4. Fix the lint issues in code or config rather than suppressing them unless suppression is clearly correct.
5. Rerun the same focused lint check after each fix until it works.
6. Keep changes minimal and avoid unrelated formatting churn.
7. Summarize the commands you ran, the files you changed, and any remaining blockers.
