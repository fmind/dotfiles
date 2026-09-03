---
name: resolve-conflicts
description: Resolve git merge or rebase conflicts by reading both sides, keeping the intent of each change, then continuing and re-running the gate. Use when a merge or rebase conflicts.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/resolve-conflicts
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Resolve Conflicts

Finish a stopped `git merge` or `git rebase` by understanding what each side meant, not by picking a side. Never `--abort` to make the problem disappear, and never "take ours" unless history shows the other change is obsolete. Branch naming lives in [feature-branch](../feature-branch/SKILL.md); committing and pushing in [git-add-commit-push](../git-add-commit-push/SKILL.md).

## Workflow

1. **Map the state**: which operation stopped, which files conflict, and which commits touch them.
   ```bash
   git status --short                    # UU both modified; AU, UA, DU, UD: added or deleted on one side
   git diff --name-only --diff-filter=U
   git log --merge --oneline -- <file>   # commits from both heads that touch the file
   ```
1. **Read both intents**: for each file, compare the working tree with the three stages and read the history behind each side (commit messages, linked pull requests and issues) before touching a hunk.
   ```bash
   git diff --base <file>     # against the common ancestor
   git diff --ours <file>     # against our side (rebase: the branch being rebased onto)
   git diff --theirs <file>   # against their side (rebase: the commit being replayed)
   git log -p $(git merge-base HEAD MERGE_HEAD)..MERGE_HEAD -- <file>   # rebase: REBASE_HEAD
   ```
1. **Resolve semantically**: write the code that satisfies both intents (renamed function plus new caller, both new tests, merged config keys). When intents are incompatible, keep the one matching the merge's goal and record the trade-off in the commit body. Do not invent new behavior, and remove every `<<<<<<<`, `=======`, `>>>>>>>` marker.
1. **Continue**: stage the resolved files and resume; repeat per commit during a rebase.
   ```bash
   git add <file>...
   git rebase --continue   # or: git merge --continue
   ```
1. **Prove it**: Run the full gate (`mise run all`); if the tree carries unrelated changes and the gate write-formats, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test` (see [mise](../mise/SKILL.md)). Fix what the merge broke before pushing.
1. **Push**: a rebased private branch needs `git push --force-with-lease`; never force-push a shared branch (`main`, or one others build on), merge into it instead.

## Gotchas

- **Show the ancestor**: `git config merge.conflictStyle zdiff3` puts the base version inside the markers so both sides' edits are visible.
- **Lockfiles and generated code**: never hand-merge `go.sum`, `uv.lock`, `pnpm-lock.yaml`, or generated files; take one side, then regenerate (`go mod tidy`, `uv lock`, `pnpm install`, `sqlc generate`).
- **Deleted on one side**: `DU` or `UD` means one side removed the file; find out why before restoring it.
- **Rebase repeats**: the same hunk can conflict on several commits; `git config rerere.enabled true` replays a recorded resolution.
- **Stop when unsure**: if intent cannot be recovered from history, ask the author instead of guessing.

## Documentation

- [git merge](https://git-scm.com/docs/git-merge#_how_to_resolve_conflicts) · [git rebase](https://git-scm.com/docs/git-rebase) · [git rerere](https://git-scm.com/docs/git-rerere)
- Adapted from [mattpocock/skills resolving-merge-conflicts](https://github.com/mattpocock/skills/blob/321658273cb1d20b76026717d027d505790106d4/skills/engineering/resolving-merge-conflicts/SKILL.md).
- Companion skills: [git-add-commit-push](../git-add-commit-push/SKILL.md) (commit and push), [repository-history](../repository-history/SKILL.md) (why a change exists), [mise](../mise/SKILL.md) (the gate).
