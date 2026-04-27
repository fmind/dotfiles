---
name: run-pre-commit-hooks
description: Run, install, autoupdate, and debug pre-commit hooks against a repo's .pre-commit-config.yaml.
---

# Run pre-commit Hooks

Operational counterpart to the `configure-pre-commit` skill. Use this when the user wants to **execute** pre-commit (run all hooks, fix failures, refresh pins, install the git hook) rather than author the config.

## When to Trigger

- The user wants to run lint/format/test hooks across the repo before committing.
- A `.pre-commit-config.yaml` exists but `git commit` doesn't fire any hooks (probably not installed).
- The user wants to update hook revisions to the latest tags.
- A pre-commit hook is failing in CI and the user wants to reproduce locally.

## Quick Reference

```bash
# Install the git hook (one-time per clone).
pre-commit install
pre-commit install --hook-type pre-push        # also fire on push
pre-commit install --hook-type commit-msg      # for conventional-commit hooks

# Run all hooks against staged files (what `git commit` would do).
pre-commit run

# Run all hooks against EVERY tracked file (initial sweep, CI).
pre-commit run --all-files

# Run a single hook by id.
pre-commit run ruff
pre-commit run ruff-format --all-files

# Bump rev: pins to latest tags.
pre-commit autoupdate

# Clean caches when something is wedged.
pre-commit clean
pre-commit gc                                  # prune unused hook envs
pre-commit uninstall                           # remove the git hook
```

## Standard Workflows

### 1. First-time setup in a fresh clone

```bash
pre-commit install
pre-commit run --all-files                     # likely modifies many files on first run
git add -A
git commit -m "chore: apply pre-commit baseline"
```

### 2. Reproduce a CI failure locally

```bash
# Match the CI invocation exactly.
pre-commit run --all-files --show-diff-on-failure --verbose
```

`--show-diff-on-failure` is what `pre-commit/action` uses; failures print the diff inline so you can apply fixes.

### 3. Fix a single hook quickly

```bash
# Identify the failing hook from `git commit` output, then:
pre-commit run <hook-id> --files path/to/file.py
```

`--files` scopes to a path list; `--all-files` runs the hook against everything.

### 4. Update revisions safely

```bash
# Bump all rev: pins to the latest tag in their upstream repos.
pre-commit autoupdate

# Run the updated hooks against the whole repo to catch behavior changes.
pre-commit run --all-files

# Commit the .pre-commit-config.yaml change separately from any code fixes it triggered.
git add .pre-commit-config.yaml
git commit -m "chore: bump pre-commit hooks"
```

### 5. Skip a hook for one commit (rare, with reason)

```bash
SKIP=ruff git commit -m "wip"
SKIP=ruff,mypy git commit -m "wip"
```

Don't make this a habit — bypassing hooks regularly defeats them.

## Diagnosing Common Failures

**"Hook failed and modified files."**
The hook ran a fixer (e.g. `ruff --fix`, `prettier`) and changed files. Re-stage and re-commit:

```bash
git add -u && git commit
```

**"Repository not found / cannot pull."**
Hook repo is unreachable. Either offline or the URL changed.

```bash
pre-commit clean && pre-commit install --install-hooks
```

**"hookid: ... not found."**
The hook id doesn't exist in the pinned rev. Likely the upstream repo renamed it.

```bash
pre-commit autoupdate
# Then update .pre-commit-config.yaml's `id:` for the renamed hook.
```

**"Executable not found."**
A `language: system` hook expects a tool that isn't installed locally. Either install the tool or switch the hook to a managed `language: python|node|rust` form.

**Hook is unbearably slow.**
Inspect with `--verbose`; common fixes:
- Scope with `files: '\.ext$'` or `types: [...]`.
- Pin to a hook env with `language_version` to avoid full re-installs.
- Use `--hook-stage=pre-push` for slow hooks (run only on push, not every commit).
- `pre-commit gc` to drop stale envs.

## CI Integration

The official action caches hook envs across runs:

```yaml
# .github/workflows/lint.yml
name: lint
on: [pull_request]
jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: "3.13" }
      - uses: pre-commit/action@v3.0.1
```

Local-equivalent invocation:

```bash
pre-commit run --all-files --show-diff-on-failure
```

## Useful Environment Variables

| Variable | Effect |
|----------|--------|
| `SKIP=hook1,hook2` | Skip listed hooks for one invocation |
| `PRE_COMMIT_HOME=/path` | Override hook env cache location (default `~/.cache/pre-commit`) |
| `PRE_COMMIT_NO_CONCURRENCY=1` | Disable parallelism (debug) |
| `PRE_COMMIT_COLOR=never` | Disable color output |

## Companion: `configure-pre-commit`

For authoring `.pre-commit-config.yaml` (choosing hooks, scoping, etc.), use the `configure-pre-commit` skill. This skill is purely operational.

## Important Notes

1. **`pre-commit install` writes to `.git/hooks/pre-commit`** — it's per-clone, not committed.
2. **`pre-commit run --all-files` is the most useful invocation** — equivalent to "what would CI catch".
3. **Hooks may modify files** (formatters); the agent should re-stage and re-run if the working tree was changed.
4. **`SKIP=...` is for emergencies**, not workflows. Document the reason in the commit message.
5. **`pre-commit gc`** safely prunes stale envs; do it occasionally to free disk space.

## Documentation

- [pre-commit home](https://pre-commit.com)
- [CLI reference](https://pre-commit.com/#cli)
- [`pre-commit/action` (GitHub Actions)](https://github.com/pre-commit/action)
- [Hooks index](https://pre-commit.com/hooks.html)
- Companion skill: `configure-pre-commit` (authoring the config).
