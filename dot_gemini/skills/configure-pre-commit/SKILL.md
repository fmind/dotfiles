---
name: configure-pre-commit
description: Guide for authoring .pre-commit-config.yaml — choosing hooks, pinning revisions, scoping by file type, fixing failures, and CI integration.
---

# Configure pre-commit (`.pre-commit-config.yaml`)

[pre-commit](https://pre-commit.com) is a multi-language framework for running checks (linters, formatters, type-checkers, secret scanners) before commits land. Hooks are sourced from versioned Git repos and isolated in their own environments — no global tool installs needed.

## Bootstrap

```bash
# Install via mise (already configured for this user).
mise use -g pipx:pre-commit

# Or with uv tool.
uv tool install pre-commit

# Generate a starter config.
pre-commit sample-config > .pre-commit-config.yaml

# Wire it into the repo.
pre-commit install                       # installs the git hook
pre-commit install --hook-type pre-push  # also fire on push (slower checks)

# Run all hooks on every file (initial pass).
pre-commit run --all-files
```

## Minimal `.pre-commit-config.yaml`

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v6.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
        args: ["--maxkb=1024"]
      - id: check-merge-conflict
      - id: check-toml
      - id: detect-private-key
```

`rev` should be a **tag** (not `master`/`main`); pre-commit caches by rev, so a moving branch defeats the cache.

## Common Stacks

### Python — Ruff + ty (Astral)

```yaml
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.7.4
    hooks:
      - id: ruff
        args: ["--fix"]
      - id: ruff-format

  # ty (Astral's type checker) doesn't ship an official pre-commit hook yet.
  # Run it as a local hook using `language: system` (requires `ty` on PATH, e.g. via uv/pipx/mise).
  - repo: local
    hooks:
      - id: ty
        name: ty type-check
        entry: ty check
        language: system
        types: [python]
        pass_filenames: false
```

### JavaScript / TypeScript — Prettier + ESLint

```yaml
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0-alpha.8
    hooks:
      - id: prettier
        types_or: [javascript, jsx, ts, tsx, json, yaml, markdown, css]

  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v9.14.0
    hooks:
      - id: eslint
        files: \.[jt]sx?$
        types: [file]
```

### Shell scripts

```yaml
  - repo: https://github.com/koalaman/shellcheck-precommit
    rev: v0.10.0
    hooks:
      - id: shellcheck

  - repo: https://github.com/scop/pre-commit-shfmt
    rev: v3.10.0-1
    hooks:
      - id: shfmt
        args: ["-i", "2", "-ci", "-bn", "-w"]
```

### Terraform

```yaml
  - repo: https://github.com/antonbabenko/pre-commit-terraform
    rev: v1.96.1
    hooks:
      - id: terraform_fmt
      - id: terraform_validate
      - id: terraform_tflint
```

### Secret scanning

```yaml
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.21.2
    hooks:
      - id: gitleaks
```

### Conventional Commits (commit-msg hook)

```yaml
  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v3.5.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
```

Then: `pre-commit install --hook-type commit-msg`.

## Scoping Hooks

| Filter | Effect |
|--------|--------|
| `files: '\\.py$'` | regex on file paths |
| `exclude: '^vendor/'` | regex of paths to skip |
| `types: [python]` | match by file type (more reliable than regex) |
| `types_or: [python, pyi]` | match any of these types |
| `stages: [pre-push]` | only on push, not every commit |
| `language_version: python3.13` | force a specific runtime for the env |

## Custom / Local Hooks

```yaml
  - repo: local
    hooks:
      - id: typecheck
        name: tsc --noEmit
        entry: bash -c 'pnpm tsc --noEmit'
        language: system
        types_or: [ts, tsx]
        pass_filenames: false

      - id: pytest-quick
        name: pytest -q -x
        entry: pytest
        language: system
        types: [python]
        args: ["-q", "-x"]
        pass_filenames: false
        stages: [pre-push]
```

`language: system` reuses the local environment (faster, but the hook is no longer fully isolated). Use it for project-specific checks tied to your venv / lockfile.

## Auto-Update Versions

```bash
pre-commit autoupdate           # bump all rev: pins to latest tags
pre-commit autoupdate --bleeding-edge   # bump to default branch (less stable)
```

Commit the updated `.pre-commit-config.yaml`; CI then runs against the new pins.

## CI Integration (recommended)

Run pre-commit in CI so checks aren't bypassed via `git commit --no-verify`:

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

The `pre-commit/action` caches hook environments between runs.

## Hooks at a Glance

| Use case | Repo |
|----------|------|
| Fix whitespace, EOF, large files | `pre-commit/pre-commit-hooks` |
| Python lint + format | `astral-sh/ruff-pre-commit` |
| Python type-check | `mypy` (`pre-commit/mirrors-mypy`) or ty as a local hook |
| JS/TS format | `pre-commit/mirrors-prettier` |
| Shell lint | `koalaman/shellcheck-precommit` + `scop/pre-commit-shfmt` |
| Terraform | `antonbabenko/pre-commit-terraform` |
| Secrets | `gitleaks/gitleaks` |
| Commit-msg style | `compilerla/conventional-pre-commit` |
| YAML | `adrienverge/yamllint` |
| Markdown | `igorshubovych/markdownlint-cli` |

## Common Workflows

**Bootstrap a project.**

```bash
pre-commit sample-config > .pre-commit-config.yaml
# Edit to add the hooks you actually want.
pre-commit install
pre-commit run --all-files          # initial pass — likely modifies many files
git add -A && git commit -m "chore: enable pre-commit"
```

**Skip a hook for one commit (rare, with reason).**

```bash
SKIP=ruff git commit -m "wip"
```

**Diagnose a slow hook.**

```bash
pre-commit run --all-files --verbose
pre-commit gc                       # prune cached hook envs
```

## Companion Skill: `run-pre-commit-hooks`

For explicit invocation of pre-commit (install, autoupdate, run-all-files), see the `run-pre-commit-hooks` skill.

## Important Notes

1. **Pin `rev:` to a tag** — `master` defeats caching and silently changes behavior.
2. **Don't use `--no-verify`** to bypass failing hooks. Fix the hook or scope it; otherwise the team-wide policy degrades.
3. **`language: system` is fast but not portable** — CI must install the same tools. Use `language: python` / `node` / etc. for portable hooks.
4. **`pre-commit run --all-files` on first install will modify many files** — commit those changes before adding more hooks.
5. **`autoupdate` should land in its own PR** so review can focus on rule changes, not unrelated edits.

## Documentation

- [pre-commit home](https://pre-commit.com)
- [Hooks index](https://pre-commit.com/hooks.html)
- [Filter options (`files`, `types`, `stages`)](https://pre-commit.com/#filtering-files-with-types)
- [`pre-commit/action` (GitHub Actions)](https://github.com/pre-commit/action)
- [`pre-commit/pre-commit-hooks` (the canonical hooks pack)](https://github.com/pre-commit/pre-commit-hooks)
