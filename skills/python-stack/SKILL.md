---
name: python-stack
description: Build typed Python projects with uv, Ruff, ty, pytest, Litestar, and Typer. Use for packages, CLIs, web apps, agents, tests, or typing.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/python-stack
  created: "2026-06-23"
  updated: "2026-09-03"
---

# Python Stack Standard

Canonical Python development: scaffolding, libraries, CLIs, Litestar web apps, and the Python API side of ADK agents. Single-file scripts belong to [python-script](../python-script/SKILL.md); the agent workflow belongs to [google-adk](../google-adk/SKILL.md).

## 1. Core & Quality Stack

- **Python**: target latest stable; use modern syntax (pattern matching, PEP 695 generics, `typing.Annotated`).
- **Dependencies**: `uv` exclusively — `uv add`, `uv run`, no manual venvs; commit `uv.lock` in applications (libraries may omit it).
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).
- **Linting and formatting**: Ruff (`ruff check --fix`, `ruff format`) with zero warnings and no `print` (`T201`); dprint for config and markup per [dprint](../dprint/SKILL.md).
- **Types**: `ty check` strict; `ty` is pre-1.0, so pin a compatible range and keep suppressions narrow and evidenced.
- **Testing**: `pytest` in `tests/` with `anyio` and an 85% branch-coverage gate; the default suite is offline, and web integration tests opt into a disposable Postgres via [conftest.py](references/conftest.py).
- **Security**: `uv audit` scans dependencies as `check:vuln` and `gitleaks` is `check:leaks`; SAST is opt-in per [opengrep](../opengrep/SKILL.md).
- **Validation and config**: Pydantic v2 and `pydantic-settings` `BaseSettings`; typed `config.py`, YAML only for cross-language needs.
- **Logging**: `structlog` — `ConsoleRenderer` locally, `JSONRenderer` in production, stdlib loggers routed through it.

## 2. Project Scaffolding Workflow

1. **Information**: define `Slug`, `Description`, `Holder/Year`, and `Package` (`Slug` with underscores) — every import path uses `Package`.
1. **Bootstrap** (agents use [google-adk](../google-adk/SKILL.md) instead):
   ```bash
   uv init --app --package --build-backend uv --vcs none --description "<description>" <slug>
   cd <slug> && uv python pin <major.minor>  # align .python-version with requires-python
   ```
1. **Manifest**: `pyproject.toml` from [pyproject.toml.template](references/pyproject.toml.template) with one profile — web keeps the Web block; CLI adds `typer` and drops the Web block and `testcontainers`; library also drops `[project.scripts]`.
1. **Config files**:
   - [mise.toml](references/mise.toml) (swap `watch` for non-web projects) and [lefthook.yml](references/lefthook.yml).
   - `dprint.json` per [dprint](../dprint/SKILL.md); `.env.example` from [env.example](references/env.example); `.gitignore` from [gitignore](references/gitignore).
   - `AGENTS.md` from [AGENTS.md](references/AGENTS.md); `LICENSE` per [project-license](../project-license/SKILL.md).
1. **Sources**: `src/<package>/__init__.py` from [init.py](references/init.py) (web), [init-cli.py](references/init-cli.py) (CLI), or [init-library.py](references/init-library.py); web and CLI add `__main__.py` from [main.py](references/main.py).
1. **Tests**: `tests/__init__.py` plus [test_smoke.py](references/test_smoke.py), then per profile:
   - Web: [test_web.py](references/test_web.py), [test_integration.py](references/test_integration.py), and root [conftest.py](references/conftest.py) (only `test:integration` starts Postgres).
   - CLI: [test_cli.py](references/test_cli.py). Library: [test_library.py](references/test_library.py).
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`; `check:leaks` scans zero commits and passes until the first commit.
1. **Finish**: `README.md` per [readme-agents](../readme-agents/SKILL.md), then `git add . && git commit -m "chore: initial commit"`.

## 3. Project Profiles

- **CLI**: a Typer app in `__init__.py` ([init-cli.py](references/init-cli.py)) exposed through `[project.scripts]`; flags, streams, and exit codes follow [cli-contracts](../cli-contracts/SKILL.md).
- **Library**: no runtime dependencies by default; omit `__main__.py` and `[project.scripts]`.
- **Data, ML, notebooks**: extension profiles on the library profile — `uv add` only the workload boundary (Polars or DuckDB, scikit-learn or a hardware-specific PyTorch/JAX, JupyterLab or Marimo) in a dependency group.
- **Scripts**: single-file tools with PEP 723 metadata go through [python-script](../python-script/SKILL.md).

## 4. Web Stack (Litestar)

- **Framework and data**: Litestar with `asyncpg` + SQLAlchemy 2 async sessions injected through `Provide`; Alembic migrations (`uv run alembic init --template async alembic`, `postgresql+asyncpg` URL).
- **Health**: `/health` is dependency-free liveness; `/ready` runs `SELECT 1` and returns 503 on `SQLAlchemyError`.
- **HTTP client**: `httpx.AsyncClient`.
- **Static assets**: self-hosted under `/static/` with SHA-256 cache busting and long-lived cache headers; no CDNs.
- **Server**: `granian` with `uvloop`, passing `Interfaces.ASGI` (the enum, so `ty` stays clean).
- **Cloud logging**: `structlog` JSON with GCP keys (`severity`, `time`, `message`, `stack_trace`), `x-cloud-trace-context` correlation, silent `/health`.

## 5. ADK Agents (Python API)

The agent workflow, `agents-cli` commands, model-pin rationale, and deployment live in [google-adk](../google-adk/SKILL.md); this section keeps the Python specifics.

- **Layout**: `app/agent.py` defines `root_agent` and its tools — plain typed functions whose signature and docstring become the JSON schema; business logic stays in modules the tools call.
- **Normalize the scaffold**: keep the generator's Python range, replace its dummy unit test, remove blanket `[tool.ty.rules]` suppressions, and fix each diagnostic at the source.
- **Offline tests**: import `root_agent`, assert its wiring (name, model, tools), and call tool functions directly; the generated `tests/integration` hits the provider and is approval-gated.

## Gotchas

- **Read installed source, never import to inspect**: `uv pip show <dist>` gives version and location, then `rg -n '^(class|def) <Symbol>\b' .venv/lib/python*/site-packages/<module>` finds the definition.
- **`uv init` Python pin**: it writes `.python-version` for whatever interpreter it resolves; run `uv python pin <major.minor>` or `uv sync` breaks.
- **`Slug` vs `Package`**: hyphenated slugs stay for the distribution, directory, image tag, and command; imports, `[project.scripts]` targets, and `python -m` use underscores.
- **`uv_build` upper bound**: keep `[build-system].requires` at least one minor ahead of the pinned `uv`, or `uv build` warns.
- **`ty` version key**: `[tool.ty.environment].python-version` takes `major.minor` only.
- **Mise dotenv**: `_.source = ".env"` in `mise.toml` loads the environment for every task.

## Documentation

- [Python](https://docs.python.org/3/) · [uv](https://docs.astral.sh/uv/) · [Ruff](https://docs.astral.sh/ruff/) · [ty](https://docs.astral.sh/ty/) · [Litestar](https://docs.litestar.dev/) · [pytest](https://docs.pytest.org/)
- Companion skills: [google-adk](../google-adk/SKILL.md) (agents), [python-script](../python-script/SKILL.md), [cli-contracts](../cli-contracts/SKILL.md), [containerize](../containerize/SKILL.md), [github-actions](../github-actions/SKILL.md), [secure](../secure/SKILL.md).
