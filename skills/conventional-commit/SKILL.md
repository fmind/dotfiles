---
name: conventional-commit
description: Write a Conventional Commits subject for the staged changes and commit them. Use when committing staged work with a typed, scoped message.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/conventional-commit
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Conventional Commit

Turn the staged changes into one Conventional Commits subject and commit them; [git-add-commit-push](../git-add-commit-push/SKILL.md) owns staging, hook healing, and pushing.

## Workflow

1. **Inspect the staged state**; if nothing is staged, say so and stop:

   ```bash
   git diff --cached --name-only   # staged files
   git diff --cached --stat        # diff stat
   git diff --cached               # full patch
   ```

1. **Read context** only when the patch alone is ambiguous: the staged files and their neighbors.
1. **Write one subject** as `<type>(<scope>): <description>`:
   - `<type>` is `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `ci`, `build`, `style`, or `revert`; `<scope>` is optional and used only when it adds signal.
   - Imperative mood (`add`, not `added`), under 72 characters in total.
   - Breaking change: append `!` after the type or scope, e.g. `feat(api)!: drop v1 endpoint`.
1. **Commit** with that exact subject and read the short hash from the output (or `git rev-parse --short HEAD`):

   ```bash
   git commit -m "<subject>"
   ```

1. **Report** only these two lines after success:

   ```text
   Subject: <subject>
   Commit: <hash>
   ```

1. **Stop on failure** (pre-commit hook, nothing staged): show the failure briefly; do not amend.

## Gotchas

- **No push**: this skill never runs `git push`; [git-add-commit-push](../git-add-commit-push/SKILL.md) does.

## Documentation

- [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- Companion skills: [feature-branch](../feature-branch/SKILL.md) (branch first), [git-add-commit-push](../git-add-commit-push/SKILL.md) (stage, commit, push), [github-pull-request](../github-pull-request/SKILL.md) (open the PR).
