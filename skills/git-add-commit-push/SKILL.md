---
name: git-add-commit-push
description: Stage, commit (Conventional Commits), and push in one flow, healing lefthook pre-commit and pre-push failures. Use when committing and pushing work end-to-end.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/git-add-commit-push
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Git Add, Commit, and Push

Stage, commit, and push in one flow, healing lefthook pre-commit and pre-push failures on the way; [conventional-commit](../conventional-commit/SKILL.md) owns the subject grammar.

## Workflow

1. **Pick the branch**: `git branch --show-current`. Direct commits to `main` are the rule for `github.com/fmind/*`; elsewhere, or for a requested PR flow, branch first with [feature-branch](../feature-branch/SKILL.md).
1. **Stage**: `git diff --cached --name-only` shows what is staged; if nothing is, `git status --short` shows the unstaged changes to add with `git add`. A clean tree ends the flow: say so and stop.
1. **Write the subject** with the [conventional-commit](../conventional-commit/SKILL.md) rules, but do not run its commit step: this skill commits and heals hook failures itself.
1. **Commit and heal pre-commit**:

   ```bash
   git commit -m "<subject>"
   ```

   On a hook failure, read its output, run `mise run format` and `mise run check`, fix type or compile errors, `git add` the touched files, and retry with the same subject until it passes or a blocker needs the user.
1. **Push and heal pre-push**:

   ```bash
   git push -u origin "$(git branch --show-current)"
   ```

   On a `mise run test` failure, read the runner output, fix the code or tests, `git add` the fix, fold it in with `git commit --amend --no-edit`, and retry until the push passes or a blocker needs the user.
1. **Report** after a successful push:

   ```text
   Subject: <subject>
   Commit: <hash>
   Status: Pushed to origin/<branch>
   ```

## Gotchas

- **Hooks are the gate**: Do not use `--no-verify` or bypass hooks unless the user explicitly asks; diagnose and fix the failure instead.
- **One commit per push**: amend hook fixes into the pending commit before pushing; never amend a commit that already reached the remote.
- **Rejected push**: branch protection on `main` means the repository wants a PR flow; switch to [feature-branch](../feature-branch/SKILL.md) and [github-pull-request](../github-pull-request/SKILL.md).

## Documentation

- [lefthook](../lefthook/SKILL.md) — the pre-commit and pre-push hooks this flow heals.
- Companion skills: [conventional-commit](../conventional-commit/SKILL.md) (subject rules), [feature-branch](../feature-branch/SKILL.md) (branch first), [github-pull-request](../github-pull-request/SKILL.md) (PR flow).
