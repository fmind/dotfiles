---
name: lefthook
description: Canonical lefthook git-hooks setup — pre-commit (format, check, secret scan) and pre-push (test), each delegating to `mise run` tasks. Use when configuring or debugging git hooks.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/lefthook
  created: 2026-07-04
  updated: 2026-07-06
---

# Lefthook Git Hooks Standard

Canonical git-hooks setup using **lefthook**. Keep hooks thin by delegating every command to a `mise run` task, so local hooks and CI run the exact same checks. Lefthook decides _when_ to run; [mise](../mise/SKILL.md) owns _what_ each command does.

## Principles

- **pre-commit** (fast): format staged files, then run the static checks and secret scan.
- **pre-push** (slower): the test suite.
- **Delegate, don't duplicate**: every command is a thin `mise run <task>`; the command name mirrors the task (`format:go` → `mise run format:go`). Tasks are owned by the language stack ([go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md)) — see the [mise skill](../mise/SKILL.md).
- **Staged formatters, whole-tree checks**: formatters take `{staged_files}` and restage their fixes (`stage_fixed: true`); `check`/`test` take no files so they always run on the whole tree — "run everything before commit/push".
- **Clean output**: suppress version headers and successful commands to keep commits quiet and distraction-free.

## Setup

1. Add lefthook using mise (Go: [go-stack](../go-stack/SKILL.md); Python: [python-stack](../python-stack/SKILL.md)).
1. Create `lefthook.yml` at the repo root (template below).
1. Install the hooks: `lefthook install` (wired into `mise run install`).

## Template

```yaml
pre-commit:
  parallel: false
  commands:
    format:dprint:
      glob: "*.{json,md,toml,yaml,yml}"
      run: mise run format:dprint {staged_files}
      stage_fixed: true
    format:<lang>: # one per language: format:go / format:python / format:templ ...
      glob: "*.<ext>"
      run: mise run format:<lang> {staged_files}
      stage_fixed: true
    check:leaks: # staged secret scan — history-mode gitleaks in `check` can't gate the incoming commit
      run: mise run check:leaks --staged
    check:
      run: mise run check
pre-push:
  commands:
    test:
      run: mise run test
```

See the reference `lefthook.yml` in [go-stack](../go-stack/references/lefthook.yml) and [python-stack](../python-stack/references/lefthook.yml).

## Gotchas

- **Thin hooks**: never inline tool commands; call `mise run <task>` so CI stays identical.
- **Ordering**: lefthook runs a hook's commands **alphabetically by name**, not in file order — so `check` sorts before `format:*` and, left unordered, runs first and fails on still-unformatted files. Set explicit `priority` (formatters low, `check` high) **with** `parallel: false` so formatters restage before `check` reads from disk.
- **Bypass**: avoid `--no-verify`; fix the failure instead — the [git-add-commit-push](../git-add-commit-push/SKILL.md) skill auto-heals hook failures.
- **Unstaged changes**: Lefthook stashes unstaged changes automatically under the hood to prevent accidental commits of unstaged changes when formatting.

## Documentation

- [Lefthook Documentation](https://lefthook.dev)
- [github-actions](../github-actions/SKILL.md) — runs the same `mise run` tasks these hooks delegate to in CI, so hooks and CI stay identical.
