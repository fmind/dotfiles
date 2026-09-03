---
name: feature-branch
description: Create and switch to a new git branch with conventional <type>/<slug> naming. Use when starting work that needs its own branch off main.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/feature-branch
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Feature Branch

Create and switch to a `<type>/<slug>` branch off `main` for the work the user described; [conventional-commit](../conventional-commit/SKILL.md) owns the commits that follow.

## Workflow

1. **Ask when the work is undescribed**: without a description, an issue reference, or a desired branch name, ask and stop.
1. **Inspect the tree**:

   ```bash
   git branch --show-current   # current branch
   git status --short          # uncommitted changes
   ```

1. **Derive the name** as `<type>/<slug>`:
   - `<type>`: a commit type from [conventional-commit](../conventional-commit/SKILL.md), usually `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, or `ci`.
   - `<slug>`: lowercase ASCII kebab-case, under 50 characters, no trailing punctuation.
1. **Reuse a valid name**: when the user's input already is a valid branch name, use it as is.
1. **Confirm the base**: off a branch other than `main` (or the default branch), warn and ask first.
1. **Confirm a dirty tree**: uncommitted changes move with the new branch; surface them and ask before continuing.
1. **Create and switch**; if the branch already exists, stop and report it:

   ```bash
   git switch -c <branch>
   ```

1. **Report** only these two lines after success:

   ```text
   Branch: <branch>
   From: <parent-branch>
   ```

## Gotchas

- **No push**: the branch stays local; [github-pull-request](../github-pull-request/SKILL.md) pushes it with `-u` when the PR is opened.

## Documentation

- [Conventional Branch](https://conventionalbranch.org/)
- Companion skills: [conventional-commit](../conventional-commit/SKILL.md) (commit on the branch), [github-pull-request](../github-pull-request/SKILL.md) (open the PR).
