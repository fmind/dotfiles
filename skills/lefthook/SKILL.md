---
name: lefthook
description: "Canonical lefthook git-hooks setup: pre-commit (format, check, secret scan) and pre-push (test), each delegating to mise run tasks. Use for git hook configuration."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/lefthook
  created: "2026-07-04"
  updated: "2026-09-03"
---

# Lefthook

Thin git hooks that delegate every command to a `mise run` task so hooks and CI run identical checks; lefthook decides _when_, [mise](../mise/SKILL.md) owns _what_.

## Workflow

1. **Install**: pin `lefthook` via mise or the stack's dev dependencies ([go-stack](../go-stack/SKILL.md), [python-stack](../python-stack/SKILL.md), [typescript-stack](../typescript-stack/SKILL.md)).
1. **Configure**: create `lefthook.yml` at the repository root from the template below; reference files live in [go-stack](../go-stack/references/lefthook.yml) and [python-stack](../python-stack/references/lefthook.yml).
1. **Activate**: `lefthook install`, wired into `mise run install`.

## Template

```yaml
output: [summary, failure, execution_out] # quiet commits: no version banner, no successful-step noise
pre-commit:
  parallel: false
  commands:
    format:dprint:
      glob: "*.{json,md,toml,yaml,yml}"
      priority: 10
      run: mise run format:dprint {staged_files}
      stage_fixed: true
    format:<lang>: # one per language: format:go / format:python / format:templ ...
      glob: "*.<ext>"
      priority: 10
      run: mise run format:<lang> {staged_files}
      stage_fixed: true
    check:leaks: # staged secret scan: history-mode gitleaks in `check` cannot see the incoming commit
      priority: 20
      run: mise run check:leaks --staged
    check:
      priority: 30
      run: mise run check
pre-push:
  commands:
    test:
      run: mise run test
```

## Principles

- **pre-commit** (fast): format staged files, then the static checks and the staged secret scan.
- **pre-push** (slower): the test suite.
- **post-commit** (optional): rebuild and redeploy a binary built from the repo's own sources, guarded to commits that touch a build input; git ignores its exit status, so it never blocks a commit.
- **Delegate, don't duplicate**: every command is `mise run <task>` and its name mirrors the task; never inline tool commands.
- **Staged formatters, whole-tree checks**: formatters take `{staged_files}` and restage fixes with `stage_fixed: true`; `check` and `test` take no files.

## Gotchas

- **Ordering**: with `parallel: false`, commands run by ascending `priority` (`10` formatters, `20` `check:leaks`, `30` `check`); commands without a priority run last in unspecified order, so set it on every command.
- **Partially staged files**: during pre-commit lefthook hides the unstaged hunks of partially staged files and restores them afterwards, so formatters only see what is being committed.
- **Bypass**: avoid `--no-verify`; fix the failure instead — [git-add-commit-push](../git-add-commit-push/SKILL.md) heals hook failures.

## Documentation

- [Lefthook](https://lefthook.dev) · [Configuration reference](https://github.com/evilmartians/lefthook/tree/master/docs/configuration)
- Companion skills: [mise](../mise/SKILL.md) (task owner), [github-actions](../github-actions/SKILL.md) (CI runs the same tasks), [gitleaks](../gitleaks/SKILL.md) (`check:leaks --staged`).
