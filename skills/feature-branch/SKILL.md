---
name: feature-branch
description: Create and switch to a new git branch with conventional `<type>/<slug>` naming. Use when starting a piece of work that needs its own branch off main.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/feature-branch
  created: 2026-06-23
  updated: 2026-07-06
---

# Create Feature Branch

Create and switch to a new feature branch for the work the user described.

## Workflow

1. If the user has not described the intended work (description, issue reference, or desired branch name), ask and stop.
1. Inspect the working tree:
   - `git branch --show-current` — current branch.
   - `git status --short` — uncommitted changes.
1. Derive a branch name in the form `<type>/<slug>` where:
   - `<type>` is one of `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `ci`.
   - `<slug>` is kebab-case, lowercase, ASCII, under 50 characters, with no trailing punctuation.
1. If the user's input already looks like a valid branch name, reuse it as-is.
1. If the current branch is not `main` (or the repo's default branch), warn briefly and ask before branching off it.
1. If the working tree has uncommitted changes, surface them and ask before continuing — the new branch will carry them.
1. Run `git switch -c <branch>` to create and check out the branch. **Do not push.** If it fails because the branch already exists, stop and report it.
1. After success, print only these two lines:

```text
Branch: <branch>
From: <parent-branch>
```

## Gotchas

- Keep the final response plain text and compact.
- Never `git push` from this skill.

## Documentation

- [Conventional Branch Specification](https://conventionalbranch.org/) — `<type>/<description>` branch naming.
- Companion skills:
  - [conventional-commit](../conventional-commit/SKILL.md) — Commit changes.
  - [github-pull-request](../github-pull-request/SKILL.md) — Open pull requests.
