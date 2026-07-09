---
name: python-stack
description: Canonical Python development stack — uv, Ruff, ty, pytest, scaffolding, Litestar web, Typer scripts, and AI agents via agents-cli. Use for any Python project, library, CLI, or agent.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/python-stack
  created: 2026-06-23
  updated: 2026-07-06
---

# Python Stack Standard

Canonical guidelines for Python development, scaffolding, CLI scripts, web applications, and AI agents.

## 1. Core & Quality Stack

- **Python**: Target latest stable. Use modern syntax (structural pattern matching, PEP 695 type parameters/aliases, parenthesized context managers, `typing.Annotated`).
- **Dependency Management**: Use `uv` exclusively. Avoid global/manual venvs. Execute tasks via `uv run`. Always commit `uv.lock` in application repositories; libraries can omit it.
- **Task Runner**: Use `mise.toml` ([mise.toml](references/mise.toml)) for the canonical tasks: `install`, `format`, `check`, `test`, `build`, and `watch`. See the [mise skill](../mise/SKILL.md).
- **Linting & Formatting**: Use **Ruff** (`ruff check --fix` and `ruff format`) for Python; enforce zero warnings/errors and ban print statements (`T201`). Clean config/markup (JSON, Markdown, TOML, YAML) with `dprint` using the configuration maintained by the [dprint skill](../dprint/SKILL.md).
- **Type Checking**: Use `ty` wrapper (`ty check`) for strict static typing. Use modern type annotations (e.g., `list[str]`, `X | Y`). Note: `ty` is pre-1.0 (fast-moving); keep it as a local check and pin a compatible range until it reaches a stable release.
- **Git Hooks**: Use `lefthook` ([lefthook.yml](references/lefthook.yml)) for pre-commit (format, check) and pre-push (test). See the [lefthook skill](../lefthook/SKILL.md).
- **Testing**: Use `pytest` in `tests/` with `anyio` (async testing) and `coverage`. Back web integration tests with a real Postgres via `testcontainers` (no mocks) — see [conftest.py](references/conftest.py).
- **Security**: Scan dependencies for known vulnerabilities with `pip-audit` (`uv run pip-audit`), wired into `mise run check` as `check:vuln`.
- **Validation & Config**: Use `Pydantic` (v2+) & `Pydantic Settings` (`BaseSettings`). Keep configs in typed Python files (e.g., `config.py`); restrict YAML to cross-language needs.
- **Logging**: Use `structlog`. Local: `ConsoleRenderer`. Production: `JSONRenderer`. Route standard library logs (SQLAlchemy, HTTPX) through `structlog` for uniform JSON outputs.

## 2. Project Scaffolding Workflow

1. **Information**: Define project `Slug`, `Description`, and `Holder/Year`.
1. **Python Pinning**: Pin target version (`major.minor` for `[tool.ty.environment].python-version`).
1. **Bootstrap**: Run in parent directory (for AI agent projects, use the `agents-cli` bootstrap workflow instead of `uv init` — see Section 5), then pin the interpreter (`uv init` writes a `.python-version` for whatever interpreter it resolves, which is often older than `requires-python` and breaks `uv sync`):
   ```bash
   uv init --app --package --build-backend uv --vcs none --description "<description>" <slug>
   cd <slug> && uv python pin <major.minor>  # align .python-version with requires-python
   ```
1. **Config Initialization**: Copy and customize:
   - `pyproject.toml` from [pyproject.toml](references/pyproject.toml)
   - `mise.toml` from [mise.toml](references/mise.toml)
   - `lefthook.yml` from [lefthook.yml](references/lefthook.yml)
   - `dprint.json` (setup as instructed in the [dprint skill](../dprint/SKILL.md))
   - `.env` & `.env.example` from [env.example](references/env.example)
   - `AGENTS.md` from [AGENTS.md](references/AGENTS.md) (see the [readme-agents skill](../readme-agents/SKILL.md))
   - `LICENSE` from [LICENSE](references/LICENSE)
   - `.gitignore` from [gitignore](references/gitignore)
1. **Scaffold Directory**:
   - Write `src/<slug>/__init__.py` using [init.py](references/init.py).
   - Write `conftest.py` from [conftest.py](references/conftest.py) (web integration wiring), plus `tests/__init__.py` and `tests/test_smoke.py` from [test_smoke.py](references/test_smoke.py).
1. **Git & Validation**:
   - Run `git init --initial-branch=main`.
   - Run verification sequence (`install` already runs `uv sync`):
     ```bash
     mise run install
     mise run format
     mise run check
     mise run test
     ```
   - Review and commit: `git add . && git commit -m "chore: initial commit"`.

## 3. Standalone Script Template

For standalone single-file CLI scripts with PEP 723 inline dependencies, use the [python-script](../python-script/SKILL.md) skill.

## 4. Web Stack & Serving Standard

1. **Web Framework**: Use `Litestar`.
1. **Database & ORM**: PostgreSQL via `asyncpg`. Use `SQLAlchemy` (v2) with `advanced-alchemy`. Manage migrations with `Alembic`. Leverage Litestar's `SQLAlchemyDTO` to auto-generate request/response serialization schemas directly from models.
1. **HTTP Client**: Use `httpx` (`AsyncClient` preferred).
1. **Static Assets (Self-Hosted)**:
   - Serve all CSS/JS/fonts locally from `/static/` (avoid external CDNs).
   - Cache-bust using SHA-256 asset content hashes (`?v=hash`).
   - Set long-term cache headers for versioned static assets; validate unversioned assets daily.
   - Preload critical head assets; lazy-load secondary scripts on interaction.
1. **Production Server**: Use `granian` with the `uvloop` engine.
1. **Structured Cloud Logging**:
   - Route server and application logs through `structlog` as JSON.
   - Map keys for GCP Stackdriver: `level` -> `severity`, `timestamp` -> `time`, `event` -> `message`, `exception` -> `stack_trace`.
   - Trace requests via `x-cloud-trace-context` header.
   - Suppress successful logs for silent routes (e.g., `/health`, `/favicon.ico`).

## 5. Agent Stack (agents-cli)

Build GCP-based AI agents with **agents-cli** (https://github.com/google/agents-cli) and **Google ADK** (`google-adk`).

1. **Bootstrap & Scaffolding**: Use `google-agents-cli` to create the project instead of standard `uv init`.
   - **For Agent Runtime deployment** (session management is handled internally, omit `--session-type`):
     ```bash
     uvx google-agents-cli create --agent-guidance-filename AGENTS.md --deployment-target agent_runtime --cicd-runner github_actions --region europe-west1 --agent adk <slug>
     ```
   - **For other deployment targets** (e.g., `cloud_run`, support `--session-type`):
     ```bash
     uvx google-agents-cli create --agent-guidance-filename AGENTS.md --deployment-target cloud_run --cicd-runner github_actions --session-type agent_platform_sessions --region europe-west1 --agent adk <slug>
     ```
2. **Layout**:
   - `app/agent.py` — defines the `root_agent` symbol and its tools. Tools are plain typed functions; ADK infers each JSON schema from the signature and docstring. Keep business logic in the library/modules and call into it from tools.
   - `app/fast_api_app.py` — FastAPI backend server for API interaction.
3. **Models & Auth**:
   - Use `gemini-3.5-flash` by default as the model.
   - Use Google Application Default Credentials (ADC) for authentication. In local development, run `gcloud auth application-default login`.
   - In `.env`, ensure `GOOGLE_GENAI_USE_VERTEXAI=true`, `GOOGLE_CLOUD_PROJECT=<project_id>`, and `GOOGLE_CLOUD_LOCATION=<region>` (e.g., `europe-west1` or `global`) are set.
4. **Development Commands**:
   - Setup project: `uvx google-agents-cli setup`
   - Install dependencies: `agents-cli install`
   - Start local playground with live reload: `agents-cli playground`
   - Run tests: `uv run pytest tests/unit tests/integration`
   - Evaluate agent: `agents-cli eval generate` followed by `agents-cli eval grade`
   - Deploy agent: `agents-cli deploy`
5. **Testing**: Import `root_agent` and assert its wiring (name, model, tools), and exercise tool functions directly — no API key, no mocks, and no web `conftest.py`/Postgres.

## Gotchas & Guidelines

- **`uv init` Python Pin**: `uv init` writes a `.python-version` for the interpreter it resolves, which can be older than `requires-python` and breaks `uv sync`. Run `uv python pin <major.minor>` right after bootstrapping.
- **`ty` Python Version**: `[tool.ty.environment].python-version` requires `major.minor` format (e.g., `"3.14"`). Do not supply patch versions.
- **Line Length**: Ruff default line length is 120 characters.
- **Alpine.js Directives**: Unpack dicts for dynamic directives (e.g., `**{"@click": "..."}`) to pass static analysis type checks.
- **CDN Restrictions**: Never reference external CDNs; serve all assets locally.
- **Mise Dotenv**: `mise` auto-loads `.env` if configured via `_.source = ".env"`.
- **agents-cli Discovery & Layout**: `agents-cli` automatically scans and manages the project structure based on `agents-cli-manifest.yaml`. Development logic goes in `app/agent.py` and the entry point uses the `root_agent` symbol. Directives are mapped via the manifest file.
- **Alembic Async Setup**:
  1. Run: `uv run alembic init --template async alembic`.
  1. Configure connection URL via `postgresql+asyncpg` in `alembic/env.py` (or load dynamically).
  1. Import declarative base metadata and assign to `target_metadata` for autogeneration support.

## Documentation

- [Python Documentation](https://docs.python.org/3/)
- [uv Documentation](https://docs.astral.sh/uv/)
- [Ruff Documentation](https://docs.astral.sh/ruff/)
- [Litestar Documentation](https://docs.litestar.dev/)
- [Google ADK Documentation](https://google.github.io/adk-docs/)
- [agents-cli Repository](https://github.com/google/agents-cli)
- Companion skills:
  - [github-actions](../github-actions/SKILL.md) — CI that runs these same `mise run` gates.
  - [security-scan](../security-scan/SKILL.md) — audit dependencies, secrets, and licenses.
  - [containerize](../containerize/SKILL.md) — package the app into a minimal, signed image.
