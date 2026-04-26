---
description: Write a conventional commit for staged changes and run git commit.
allowed-tools: Bash(git diff:*), Bash(git commit:*)
---

Create a conventional commit for the staged changes in this repository.

Staged files:
!`git diff --cached --name-only`

Diff stat:
!`git diff --cached --stat`

Patch:
```diff
!`git diff --cached`
```

Requirements:
1. If nothing is staged, say that and stop.
2. Read the staged files or nearby context if the patch is ambiguous.
3. Write one conventional commit subject in imperative mood, under 72 characters.
4. Prefer a precise scope when it helps.
5. Use the shell command tool to run `git commit -m "<subject>"` with that exact subject.
6. After success, print only these two lines:
Subject: <subject>
Commit: <hash>
7. If the commit fails, show the failure briefly and stop.

Keep the final response plain text and compact.
