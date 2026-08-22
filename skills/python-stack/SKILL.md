---
name: python-stack
description: Canonical Python stack with uv, Ruff, ty, pytest, Litestar, Typer, and dependency source. Use for Python projects, libraries, CLIs, agents, packaging, tests, typing, linting, or verifying APIs without importing them.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/python-stack
  created: 2026-06-23
  updated: 2026-08-02
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
- **Testing**: Use `pytest` in `tests/` with `anyio` and an 85% branch-coverage gate. Keep the default suite offline; web integration tests opt into a real disposable Postgres via [conftest.py](references/conftest.py) and [test_integration.py](references/test_integration.py).
- **Security**: Scan dependencies for known vulnerabilities with `pip-audit` (`uv run pip-audit`), wired into `mise run check` as `check:vuln`.
- **Validation & Config**: Use `Pydantic` (v2+) & `Pydantic Settings` (`BaseSettings`). Keep configs in typed Python files (e.g., `config.py`); restrict YAML to cross-language needs.
- **Logging**: Use `structlog`. Local: `ConsoleRenderer`. Production: `JSONRenderer`. Route standard library logs (SQLAlchemy, HTTPX) through `structlog` for uniform JSON outputs.

## Exact Dependency Source Review

Resolve a symbol from the uv-selected environment with [resolve_source.py](scripts/resolve_source.py) before proposing a dependency-specific fix:

```bash
python3 ~/.agents/skills/python-stack/scripts/resolve_source.py <distribution> <symbol> --project <project> [--module <import-module>]
```

The resolver's offline contract suite is [resolve_source_test.py](scripts/resolve_source_test.py).

- Keep inspection read-only: the resolver parses `*.dist-info` metadata and Python ASTs without importing or executing dependency code.
- Treat an ambiguous environment, duplicate installed version, missing or stale source, or ambiguous symbol as an actionable error. Installed generated Python remains valid runtime evidence; inspect it read-only like other installed source. Pass `--environment` only after confirming the intended uv environment.
- Editable installs resolve through their local `direct_url.json`; normal installs remain confined to the selected environment's `site-packages`.
- Consume the versioned JSON result: `language`, `dependency`, `version`, `source_path`, `defining_file`, `symbol`, bounded `excerpt`, and `provenance`. This shape is compatible with the Go resolver, while environment semantics remain owned here.

## 2. Project Scaffolding Workflow

1. **Information**: Define project `Slug`, `Description`, and `Holder/Year`. Derive `Package` — `Slug` with hyphens replaced by underscores — whenever `Slug` contains a hyphen; `uv init --package` already names the source directory this way, and every Python import path (`[project.scripts]` target, `pdoc`/`granian` module args, test imports) must use `Package`, not the hyphenated `Slug`.
1. **Python Pinning**: Pin target version (`major.minor` for `[tool.ty.environment].python-version`).
1. **Bootstrap**: Run in parent directory (for AI agent projects, use the `agents-cli` bootstrap workflow instead of `uv init` — see Section 5), then pin the interpreter (`uv init` writes a `.python-version` for whatever interpreter it resolves, which is often older than `requires-python` and breaks `uv sync`):
   ```bash
   uv init --app --package --build-backend uv --vcs none --description "<description>" <slug>
   cd <slug> && uv python pin <major.minor>  # align .python-version with requires-python
   ```
1. **Config Initialization**: Copy and customize:
   - `pyproject.toml` from [pyproject.toml.template](references/pyproject.toml.template), then apply exactly one runtime profile. **Web** keeps the application and Web dependency blocks. **CLI: add `typer`** with `uv add typer`, remove the Web block and `testcontainers`, and keep `[project.scripts]`. **Library: remove `[project.scripts]`**, remove all unused application/Web dependencies and `testcontainers`, and start with no runtime dependencies unless its API needs them. Agent projects use their generated manifest instead.
   - `mise.toml` from [mise.toml](references/mise.toml) — for non-web projects, swap the web-only `watch` task (see its inline comment).
   - `lefthook.yml` from [lefthook.yml](references/lefthook.yml)
   - `dprint.json` (setup as instructed in the [dprint skill](../dprint/SKILL.md))
   - `.env` & `.env.example` from [env.example](references/env.example)
   - `AGENTS.md` from [AGENTS.md](references/AGENTS.md) (see the [readme-agents skill](../readme-agents/SKILL.md))
   - `LICENSE` from [LICENSE](references/LICENSE)
   - `.gitignore` from [gitignore](references/gitignore)
1. **Scaffold Directory**:
   - Write `src/<package>/__init__.py` from the selected profile: **web** [init.py](references/init.py), **CLI** [init-cli.py](references/init-cli.py), or **library** [init-library.py](references/init-library.py).
   - Web and CLI projects write `src/<package>/__main__.py` from [main.py](references/main.py), so `python -m <package>` and containers target the import package rather than a possibly hyphenated distribution name. Libraries omit it.
   - Write `tests/__init__.py` and the selected tests: **web** uses [test_smoke.py](references/test_smoke.py), [test_web.py](references/test_web.py), [test_integration.py](references/test_integration.py), and root [conftest.py](references/conftest.py); **CLI** uses `test_smoke.py` plus [test_cli.py](references/test_cli.py); **library** uses [test_library.py](references/test_library.py). Only `mise run test:integration` starts Postgres.
1. **Git & Validation**:
   - Run `git init --initial-branch=main`.
   - Run verification sequence (`install` already runs `uv sync`):
     ```bash
     mise run install
     mise run format
     mise run check
     mise run test
     ```
     `check:leaks` scans the working tree, including a fresh repository with no Git history.
   - Review and commit: `git add . && git commit -m "chore: initial commit"`.

## 3. Standalone Script Template

For standalone single-file CLI scripts with PEP 723 inline dependencies, use the [python-script](../python-script/SKILL.md) skill.

### Data, ML, and Notebook Extensions

Data, ML, and notebooks are extension profiles, not bundled base dependencies. Start with the library profile, keep reusable logic in `src/` with deterministic tests, and add only the workload boundary with `uv add`: Polars/PyArrow/DuckDB for analytical data, scikit-learn or a hardware-appropriate PyTorch/JAX install for ML, and JupyterLab or Marimo for notebooks. Follow vendor installation guidance for accelerator-specific wheels, put interactive tools in a dependency group, and commit the application lockfile; do not make every Python project pay their dependency, platform, and security cost.

## 4. Web Stack & Serving Standard

1. **Web Framework**: Use `Litestar`.
1. **Database & ORM**: PostgreSQL via `asyncpg`, direct SQLAlchemy 2 async sessions injected with Litestar `Provide`, and Alembic migrations. Keep the session factory behind the application factory so unit tests remain offline and every application lifespan disposes its engine.
1. **Health Endpoints**: Keep `/health` dependency-free for process liveness. Use `/ready` for a database `SELECT 1`; return 503 on `SQLAlchemyError` so orchestration stops routing traffic while dependencies are unavailable.
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
1. **Layout**:
   - `app/agent.py` — defines the `root_agent` symbol and its tools. Tools are plain typed functions; ADK infers each JSON schema from the signature and docstring. Keep business logic in the library/modules and call into it from tools.
   - `app/fast_api_app.py` — FastAPI backend server for API interaction.
1. **Models & Auth**:
   - Pin the current Flash generation by default (e.g. `gemini-3.7-flash`, matching `agents-cli`'s own scaffold) rather than the `-latest` alias — Vertex AI's `-latest` resolution has documented version-ambiguity and hot-swap quality regressions. Check `agents-cli create --agent adk`'s generated default or the [Gemini models list](https://ai.google.dev/gemini-api/docs/models) and bump to the newest Flash generation when one ships; only pin an exact dated snapshot when you need strict reproducibility.
   - Use Google Application Default Credentials (ADC) for authentication. In local development, run `gcloud auth application-default login`.
   - In `.env`, ensure `GOOGLE_GENAI_USE_VERTEXAI=true`, `GOOGLE_CLOUD_PROJECT=<project_id>`, and `GOOGLE_CLOUD_LOCATION=<region>` (e.g., `europe-west1` or `global`) are set.
1. **Development Commands**:
   - Install dependencies: `uvx google-agents-cli install --locked`
   - Start local playground with live reload: `uvx google-agents-cli playground`
   - Run the default offline suite: `uv run pytest tests/unit`
   - With explicit approval for credentials, model usage, and cost, run generated live integration tests: `uv run pytest tests/integration`
   - With the same explicit approval, evaluate: `uvx google-agents-cli eval generate` followed by `uvx google-agents-cli eval grade`
   - Deploy only after explicit approval: `uvx google-agents-cli deploy`
1. **Post-generation normalization**: The generator owns its supported Python range, which can lag the latest interpreter. Preserve that compatible range, replace its dummy unit test, remove generated blanket `[tool.ty.rules]` suppressions, and fix each remaining diagnostic at the source or with the narrowest evidenced line-level exception.
1. **Offline Testing**: Import `root_agent`, assert its wiring (name, model, tools), and exercise tool functions directly. Keep this default path free of API keys, live model calls, Postgres, and web fixtures; generated `tests/integration` exercises the provider and is approval-gated.

## Gotchas & Guidelines

- **`uv init` Python Pin**: `uv init` writes a `.python-version` for the interpreter it resolves, which can be older than `requires-python` and breaks `uv sync`. Run `uv python pin <major.minor>` right after bootstrapping.
- **`Slug` vs `Package` in Import Paths**: When `Slug` contains a hyphen (e.g. `my-cool-app`), substituting it literally into a Python import position — `[project.scripts]`'s right-hand side, `pdoc`/`granian` module arguments, imports, or `python -m` — fails. Use `Package` (underscores) there; `Slug` stays correct for the distribution name, directory, Docker tag, and installed console-script command.
- **`uv_build` Upper Bound**: Keep `[build-system].requires`'s `uv_build` upper bound at least one minor ahead of the pinned `uv` tool version — `uv sync`/`uv build` warns (and a stricter resolver would fail) when the installed `uv_build` version falls outside the declared range.
- **`ty` Python Version**: `[tool.ty.environment].python-version` requires `major.minor` format (e.g., `"3.14"`). Do not supply patch versions.
- **Line Length**: Ruff default line length is 120 characters.
- **`ty` Scope**: The scaffold ships with no blanket `[tool.ty.rules]` ignores. Because `ty` is pre-1.0, pin its compatible range and use only evidenced, narrow suppressions; never carry a generator's blanket unresolved or invalid-type ignores forward.
- **Granian Interface**: Pass the `Interfaces.ASGI` enum (`from granian.constants import Interfaces`), not the `"asgi"` string, so `ty` stays clean without an ignore.
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
