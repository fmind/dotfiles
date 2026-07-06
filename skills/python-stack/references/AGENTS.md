# AGENTS.md (Project)

Context and rules for AI agents working in this repository. Humans should start with `README.md`.

## Project overview

- **Name**: <slug>
- **Description**: <1-2 sentences on what this project does.>
- **Language**: Python 3.14+ (`pyproject.toml`), managed with `uv`.

## Setup & core commands

All work goes through `mise` (see `mise.toml`); git hooks and CI call the same tasks.

- Install: `mise run install` — sync the virtualenv (`uv sync`) and install git hooks.
- Format: `mise run format` — `ruff` (import sort + format) and `dprint`.
- Check: `mise run check` — `ruff` lint, `ty` types, `pip-audit`, `dprint check`, `gitleaks`, `pyproject` validation.
- Test: `mise run test` — `pytest` with coverage.
- Build: `mise run build` — `uv build` (wheel + sdist).
- Watch: `mise run watch` — live reload for local development (web: `granian`; agent: `agents-cli playground`).

## Definition of done

A change is complete only when, locally, `mise run format` is clean, `mise run check` reports no findings, `mise run test` is green, and new or changed behavior has a test. Fix root causes — never weaken an assertion, add a skip/`xfail`, loosen a type, or suppress a lint error to force a green result.

## Conventions & idioms

- **Errors with context**: raise specific exceptions and chain with `raise ... from err`; never use a bare `except` or silently swallow errors.
- **Config from env**: parse environment into a typed `pydantic-settings` `BaseSettings`; validate and fail fast on startup. No hardcoded values — lift constants into config.
- **Typing**: modern annotations (`list[str]`, `X | Y`, `typing.Annotated`); keep `ty check` clean. Parse external input into trusted types (Pydantic models, enums) at the boundary.
- **Logging**: `structlog` — `ConsoleRenderer` in development, `JSONRenderer` in production.
- **Async**: prefer `httpx.AsyncClient`; test async paths with `anyio`.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`); no attribution in commit messages.

## Repository layout

- `src/<slug>/__init__.py` — package entry point: the Litestar `app` (web) or the `root_agent` re-export (agent), plus `__version__`.
- `src/<slug>/agent.py` — (agent) `root_agent` definition and its typed function tools.
- `tests/` — `pytest` suite: `conftest.py` fixtures (e.g. `testcontainers`) and `test_*.py`.
- `pyproject.toml` — dependencies and `ruff`/`ty`/`pytest` config; `mise.toml` — task runner and pinned tools; `lefthook.yml` — git hooks; `dprint.json` — config/markup formatter.
- `.env` / `.env.example` — environment configuration (never commit secrets).
- Standalone CLI tools are single-file scripts with PEP 723 inline dependencies, run via `uv run <script>.py`.
