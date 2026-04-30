---
name: create-uv-project
description: Bootstrap a Python project managed by uv — pyproject.toml, uv.lock, optional package layout, dev tools, and CI integration.
---

# Create uv Project

[`uv`](https://docs.astral.sh/uv) (by Astral) is the fast all-in-one Python tool. This skill covers scaffolding a **project** (vs the `create-python-script` skill, which targets standalone PEP 723 scripts). For day-to-day CLI usage, see `use-uv-cli`.

## When to Trigger

- The user wants to start a new Python project (CLI app, library, service, ML pipeline) and needs `pyproject.toml` + `uv.lock`.
- The user mentions migrating a project from `pip` / `poetry` / `pipenv` / `setup.py` to `uv`.
- The user wants a managed venv tied to a pinned Python version.

## Bootstrap

```bash
# Create a project directory + minimal pyproject + main.py.
uv init my-app
cd my-app

# Or: app + library layout (src/-layout — preferred for distributable libs).
uv init --lib my-lib
uv init --app my-app

# Pin Python.
uv python pin 3.13          # writes .python-version

# Add deps.
uv add "fastapi>=0.110" "uvicorn[standard]>=0.30"
uv add --dev pytest pytest-asyncio ruff ty

# Run something inside the project venv.
uv run python -V
uv run pytest
```

## Resulting Layout

### `--app` (default)

```text
my-app/
├── .python-version
├── README.md
├── pyproject.toml
├── uv.lock
└── main.py
```

### `--lib` (distributable library)

```text
my-lib/
├── .python-version
├── README.md
├── pyproject.toml
├── uv.lock
└── src/
    └── my_lib/
        ├── __init__.py
        └── py.typed              # if you ship type hints
```

## Recommended `pyproject.toml`

```toml
[project]
name = "my-app"
version = "0.1.0"
description = "What this does."
readme = "README.md"
requires-python = ">=3.13"
authors = [
  { name = "Médéric Hurier", email = "mederic.hurier@fmind.dev" },
]
license = { text = "MIT" }
classifiers = [
  "License :: OSI Approved :: MIT License",
  "Programming Language :: Python :: 3.13",
]
dependencies = [
  "fastapi>=0.110",
  "uvicorn[standard]>=0.30",
]

[project.optional-dependencies]
test = ["pytest>=8", "pytest-asyncio>=0.23"]

[project.scripts]
my-app = "my_app.cli:main"          # console script entry point

[build-system]
requires = ["hatchling>=1.21"]
build-backend = "hatchling.build"

[tool.uv]
dev-dependencies = [
  "pytest>=8",
  "ruff>=0.7",
  "ty>=0.0.1",
]

[tool.ruff]
line-length = 100
target-version = "py313"
[tool.ruff.lint]
select = ["E", "F", "W", "I", "UP", "B", "SIM", "RUF"]
[tool.ruff.format]
quote-style = "double"

[tool.pytest.ini_options]
addopts = "-q -ra"
testpaths = ["tests"]
```

## Dev Tooling Bundle

```bash
# Lint + format.
uv add --dev ruff
# Type-check (pick one).
uv add --dev ty                # Astral, fast
# uv add --dev mypy             # mature alternative
# Tests.
uv add --dev pytest pytest-asyncio pytest-cov
# Run.
uv run ruff check . && uv run ruff format .
uv run ty check
uv run pytest --cov
```

## Workspaces (monorepo)

```toml
# Root pyproject.toml
[tool.uv.workspace]
members = ["packages/*"]
```

```bash
# At root:
uv sync                              # installs all workspace members
uv run --package my-pkg pytest       # run command in a specific package
```

Each subpackage has its own `pyproject.toml` and shares the root `uv.lock`.

## CLI Entry Points

`pyproject.toml`:

```toml
[project.scripts]
my-app = "my_app.cli:main"
```

```python
# src/my_app/cli.py
import typer
app = typer.Typer()

@app.command()
def hello(name: str = "world") -> None:
    print(f"Hello {name}!")

def main() -> None:
    app()
```

After `uv sync`, the entry point is on `$PATH` inside the venv:

```bash
uv run my-app hello --name=Médéric
```

## Locking & Reproducibility

```bash
uv lock                         # update uv.lock
uv sync --frozen                # error if lockfile is stale (CI)
uv sync --all-extras --dev      # install everything for local dev
uv export --format requirements-txt > requirements.txt   # for downstream tools
```

Commit `uv.lock`. CI runs `uv sync --frozen --all-extras --dev` for deterministic builds.

## Build & Publish

```bash
uv build                        # creates dist/*.whl + dist/*.tar.gz
uv publish                      # to PyPI (requires UV_PUBLISH_TOKEN)
```

Configure trusted publishing on PyPI for token-less CI uploads.

## CI (GitHub Actions example)

```yaml
name: ci
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
        with: { version: "latest" }
      - run: uv sync --frozen --all-extras --dev
      - run: uv run ruff check .
      - run: uv run ty check
      - run: uv run pytest --cov
```

## Common Workflows

**Bootstrap a CLI tool.**

```bash
uv init --lib my-tool && cd my-tool
uv python pin 3.13
uv add typer rich
uv add --dev pytest
# Add `[project.scripts] my-tool = "my_tool.cli:main"` to pyproject.toml
mkdir -p src/my_tool && touch src/my_tool/__init__.py
# Edit src/my_tool/cli.py with the Typer app.
uv run my-tool --help
```

**Migrate from pip / poetry.**

```bash
# From requirements.txt.
uv init my-app && cd my-app
uv add -r ../requirements.txt
uv lock

# From poetry (pyproject.toml exists).
uv migrate                      # interactive (where supported)
# Otherwise: rewrite [tool.poetry] → [project] manually, then `uv lock`.
```

**Add a script entry point and publish.**

1. Edit `[project.scripts]` block.
2. `uv build`
3. `uv publish` (with a configured PyPI token).

## Important Notes

1. **Commit `uv.lock` and `.python-version`** — they are the contract.
2. **Use `--app` vs `--lib`** intentionally — `--lib` produces an installable wheel and tries to import from `src/`; `--app` is for non-distributable runners.
3. **Don't activate the venv manually** — `uv run <cmd>` always runs inside `.venv/`. Activation is for interactive shell only.
4. **Build backend is `hatchling` by default** — fine for most cases. Switch to `pdm-backend` / `setuptools` only if you need their specific features.
5. **`requires-python` flows into `uv.lock`** — pinning it tightly (e.g. `>=3.13`) keeps lockfile resolution snappy.
6. **For inline scripts, use `create-python-script` instead** — PEP 723 metadata is simpler than a project for one-shots.

## Documentation

- [uv home](https://docs.astral.sh/uv)
- [Project guide](https://docs.astral.sh/uv/guides/projects/)
- [`pyproject.toml` reference](https://docs.astral.sh/uv/concepts/projects/config/)
- [Workspaces](https://docs.astral.sh/uv/concepts/workspaces/)
- [Dev dependencies](https://docs.astral.sh/uv/concepts/projects/dependencies/)
- [Build & publish](https://docs.astral.sh/uv/guides/publish/)
- [Astral's `uv` skill (Claude Code)](https://github.com/astral-sh/claude-code-plugins/blob/main/plugins/astral/skills/uv/SKILL.md)
- Companion skills: `create-python-script`, `use-uv-cli`.
