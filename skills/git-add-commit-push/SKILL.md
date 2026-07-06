---
name: git-add-commit-push
description: Stage, commit (Conventional Commits), and push in one flow, auto-healing lefthook pre-commit and pre-push failures. Use when committing and pushing work end-to-end.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/git-add-commit-push
  created: 2026-06-23
  updated: 2026-07-06
---

# Git Add, Commit, and Push

Wraps the commit-subject rules from `conventional-commit`, then stages, commits, and pushes — auto-healing lefthook pre-commit/pre-push failures along the way.

## Workflow

1. **Check Working Directory & Staging State**:
   - Run `git branch --show-current`; if on the default branch (`main`/`master`), stop and suggest branching first via the [feature-branch](../feature-branch/SKILL.md) skill (or confirm with the user) before committing and pushing.
   - Check if there are staged changes: `git diff --cached --name-only`.
   - If nothing is staged, check if there are unstaged changes: `git status --short`.
   - If there are unstaged changes, stage them using `git add` to prepare for commit.
   - If the working tree is completely clean and nothing is staged, say so and stop.

1. **Generate Commit Message**:
   - Craft the commit subject using the naming rules from the [conventional-commit](../conventional-commit/SKILL.md) skill (`<type>(<scope>): <description>`, imperative mood, under 72 chars). Do NOT run its commit/stop-on-failure steps — this skill performs the commit and auto-heals hook failures below.

1. **Commit Stage (Pre-Commit Hook Auto-Healing)**:
   - Run `git commit -m "<subject>"` to commit the staged changes.
   - If the commit fails (pre-commit hook failure):
     - Read the hook output carefully to diagnose the issues.
     - **Formatting/Linting**: Run the project's `mise run format` and `mise run check` tasks.
     - **Type checking / Tests**: Run type verification checks or compiler checks to resolve error states.
     - Once the issues are resolved, stage the affected files (`git add`) and retry the commit with the same message.
     - Repeat this troubleshooting loop until the commit succeeds, or report the blocker if manual intervention is required.

1. **Push Stage (Pre-Push Hook Auto-Healing)**:
   - Identify the current branch using `git branch --show-current`.
   - Run `git push` (remote tracking is set up automatically via `push.autoSetupRemote`).
   - If the push fails due to a pre-push hook failure (the `test` hook running `mise run test` — `pytest`/`gotestsum`):
     - Read the test runner output to identify the failing tests.
     - Debug and fix the bugs in the code or tests.
     - Stage the fixes (`git add`), amend the commit (`git commit --amend --no-edit`), and retry the push.
     - Repeat until the push succeeds, or report the remaining failures if manual intervention is required.

1. **Output**:
   - After a successful push, print the final status in a compact format:

     ```text
     Subject: <subject>
     Commit: <hash>
     Status: Pushed to origin/<branch>
     ```

## Gotchas

- Ensure all git hook failures are fully diagnosed and resolved natively. Do not use `--no-verify` or bypass hooks unless explicitly requested by the user.
- Keep the commit history clean by using `git commit --amend` to combine hook fixes into the main commit before pushing.

## Documentation

- [conventional-commit](../conventional-commit/SKILL.md) — commit subject naming rules.
