---
name: use-uv-cli
description: Guide for using uv (Astral's fast Python package manager) — projects, scripts, virtualenvs, dependency mgmt, tool installs, and Python version pinning.
---

# Use uv CLI

[uv](https://docs.astral.sh/uv) is Astral's fast, all-in-one Python tool — replaces `pip`, `pip-tools`, `pipx`, `pyenv`, `virtualenv`, `poetry`, and most `setup.py` workflows.

> Astral also publishes an official Claude-Code-flavored `uv` skill at [`astral-sh/claude-code-plugins`](https://github.com/astral-sh/claude-code-plugins/blob/main/plugins/astral/skills/uv/SKILL.md). For Gemini CLI users on this dotfile setup, this skill provides the equivalent CLI guidance.

## Install

```bash
# Recommended (also installs uv-managed Python).
curl -LsSf https://astral.sh/uv/install.sh | sh

# Or via mise (already configured for this user).
mise use -g uv@latest

# Sanity check.
uv --version
uv help
```

## Standalone Scripts (PEP 723)

```bash
# Run a script with inline metadata (no project required).
uv run script.py

# Add a dependency to a script's PEP 723 block.
uv add --script script.py "rich>=13" "typer>=0.12"

# Pin a Python version inline.
uv add --script script.py --python ">=3.13"
```

A typical script header (matches `create-python-script` skill):

```python
#!/usr/bin/env -S uv run --quiet --script
# /// script
# requires-python = ">=3.13"
# dependencies = ["rich>=13", "typer>=0.12"]
# ///
```

## Projects (`pyproject.toml` + `uv.lock`)

```bash
# Bootstrap.
uv init my-app             # creates pyproject.toml + .python-version + main.py
cd my-app

# Add / remove deps.
uv add "fastapi>=0.110" "uvicorn[standard]"
uv add --dev pytest ruff
uv remove uvicorn

# Lock + sync (idempotent).
uv lock                    # update uv.lock
uv sync                    # install from uv.lock into .venv

# Run a command in the project's venv.
uv run pytest
uv run python -m my_app
uv run uvicorn my_app:app --reload
```

## Python Version Management

```bash
# Install a specific Python.
uv python install 3.13

# Pin the project to a Python.
uv python pin 3.13         # writes .python-version

# List installed pythons.
uv python list
uv python find             # show the resolved binary
```

## Tool Installs (pipx replacement)

```bash
# Install a CLI tool into an isolated venv on $PATH.
uv tool install ruff
uv tool install black
uv tool list

# Run a tool ephemerally without installing.
uv tool run ruff check .
uvx ruff check .           # shorthand alias for `uv tool run`

# Update / uninstall.
uv tool upgrade --all
uv tool uninstall ruff
```

## Workspaces (monorepo)

```toml
# pyproject.toml at repo root
[tool.uv.workspace]
members = ["packages/*"]
```

```bash
uv sync                    # installs all workspace members
uv run --package my-pkg pytest
```

## Lockfile & Reproducibility

- `uv.lock` is the single source of truth — commit it.
- `uv sync --frozen` errors out if the lockfile is stale (use in CI).
- `uv export --format requirements-txt > requirements.txt` for downstream tools that don't speak `uv.lock`.

## Common Workflows

**Bootstrap a new project.**

```bash
uv init my-service && cd my-service
uv python pin 3.13
uv add fastapi "uvicorn[standard]"
uv add --dev pytest ruff
uv run pytest
```

**Reproduce CI locally.**

```bash
uv sync --frozen --all-extras --dev
uv run pytest
```

**Migrate from pip / poetry / pipenv.**

```bash
# Convert a Poetry / Pipenv / Setuptools project to uv (third-party tool).
uvx migrate-to-uv

# Or import requirements.txt directly into a uv project.
uv add -r requirements.txt
```

`migrate-to-uv` is published separately (`mkniewallner/migrate-to-uv`); `uv` itself has no built-in `migrate` subcommand.

## Important Notes

1. **`uv.lock` is universal across platforms** — commit it; CI uses it for deterministic installs.
2. **`uv run` always runs inside the project venv** (created on demand at `.venv/`) — don't activate it manually unless you have a reason.
3. **`uv tool` installs are isolated** — they don't pollute the project's deps.
4. **`uvx` is a shorthand** for `uv tool run`; both run the tool ephemerally without `tool install`.
5. **For inline-script deps**, prefer PEP 723 metadata blocks over project files — keeps standalone scripts standalone.

## Documentation

- [uv home](https://docs.astral.sh/uv)
- [Projects](https://docs.astral.sh/uv/concepts/projects/)
- [Standalone scripts (PEP 723)](https://docs.astral.sh/uv/guides/scripts/)
- [Tools (pipx replacement)](https://docs.astral.sh/uv/guides/tools/)
- [Python version management](https://docs.astral.sh/uv/concepts/python-versions/)
- [Workspaces](https://docs.astral.sh/uv/concepts/workspaces/)
- [Astral's `uv` skill (Claude Code)](https://github.com/astral-sh/claude-code-plugins/blob/main/plugins/astral/skills/uv/SKILL.md)
- [GitHub: astral-sh/uv](https://github.com/astral-sh/uv)
