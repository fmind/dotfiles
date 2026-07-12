---
name: conventional-commit
description: Write a Conventional Commits subject for the staged changes and commit them. Use when committing staged work with a typed, scoped message.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/conventional-commit
  created: 2026-06-23
  updated: 2026-07-09
---

# Create Conventional Commit

Create a Conventional Commits message for the staged changes in this repository and commit them.

## Workflow

1. Inspect the staged state:
   - `git diff --cached --name-only` — staged files.
   - `git diff --cached --stat` — diff stat.
   - `git diff --cached` — full patch.
1. If nothing is staged, say so and stop.
1. Read the staged files or nearby context if the patch is ambiguous.
1. Write **one** Conventional Commits subject:
   - Form: `<type>(<scope>): <description>` where `<type>` is `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `ci`, `build`, `style`, or `revert`. `<scope>` is optional.
   - Imperative mood (`add`, not `added`).
   - Under 72 characters total.
   - Precise scope when it helps; omit if it adds no signal.
   - For a breaking change, append `!` after the type/scope, e.g. `feat(api)!: drop v1 endpoint`.
1. Run `git commit -m "<subject>"` with that exact subject, then read the short hash from the commit output (or `git rev-parse --short HEAD`).
1. After success, print only these two lines:

```text
Subject: <subject>
Commit: <hash>
```

1. If the commit fails (pre-commit hook, no staged changes), show the failure briefly and stop. **Do not amend.**

## Gotchas

- Keep the final response plain text and compact.
- Never run `git push` from this skill.

## Documentation

- [github-pull-request](../github-pull-request/SKILL.md) — open or update a PR after the branch has commits.
