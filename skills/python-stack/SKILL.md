---
name: python-stack
description: Canonical Python development stack — uv, Ruff, ty, pytest, scaffolding, Litestar web, Typer scripts, AI agents, and exact local dependency source review. Use for any Python project, library, CLI, or agent, including confirmation of installed dependency APIs without importing their code.
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
- **Testing**: Use `pytest` in `tests/` with `anyio` (async testing) and `coverage`. Back web integration tests with a real Postgres via `testcontainers` (no mocks) — see [conftest.py](references/conftest.py).
- **Security**: Scan dependencies for known vulnerabilities with `pip-audit` (`uv run pip-audit`), wired into `mise run check` as `check:vuln`.
- **Validation & Config**: Use `Pydantic` (v2+) & `Pydantic Settings` (`BaseSettings`). Keep configs in typed Python files (e.g., `config.py`); restrict YAML to cross-language needs.
- **Logging**: Use `structlog`. Local: `ConsoleRenderer`. Production: `JSONRenderer`. Route standard library logs (SQLAlchemy, HTTPX) through `structlog` for uniform JSON outputs.

## Exact Dependency Source Review

Resolve a symbol from the uv-selected environment before proposing a dependency-specific fix:

```bash
python3 ~/.agents/skills/python-stack/scripts/resolve_source.py <distribution> <symbol> --project <project> [--module <import-module>]
```

The resolver's offline contract suite is [resolve_source_test.py](scripts/resolve_source_test.py).

- Keep inspection read-only: the resolver parses `*.dist-info` metadata and Python ASTs without importing or executing dependency code.
- Treat an ambiguous environment, duplicate installed version, missing or stale source, generated file, or ambiguous symbol as an actionable error. Pass `--environment` only after confirming the intended uv environment.
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
   - `pyproject.toml` from [pyproject.toml.template](references/pyproject.toml.template) — `dependencies` are grouped **Core** (every project type) and **Web (Litestar)**; delete the Web block and `testcontainers` for library/CLI/agent projects.
   - `mise.toml` from [mise.toml](references/mise.toml) — for non-web projects, swap the web-only `watch` task (see its inline comment).
   - `lefthook.yml` from [lefthook.yml](references/lefthook.yml)
   - `dprint.json` (setup as instructed in the [dprint skill](../dprint/SKILL.md))
   - `.env` & `.env.example` from [env.example](references/env.example)
   - `AGENTS.md` from [AGENTS.md](references/AGENTS.md) (see the [readme-agents skill](../readme-agents/SKILL.md))
   - `LICENSE` from [LICENSE](references/LICENSE)
   - `.gitignore` from [gitignore](references/gitignore)
1. **Scaffold Directory**:
   - Write `src/<package>/__init__.py` — **web**: [init.py](references/init.py) (the Litestar `app`); **library/CLI**: [init-cli.py](references/init-cli.py) (minimal package: `__version__`, env-aware `structlog`, `main()`).
   - Write `tests/__init__.py` and `tests/test_smoke.py` from [test_smoke.py](references/test_smoke.py). **Web** projects also add `conftest.py` from [conftest.py](references/conftest.py) (Postgres `testcontainers` wiring); library/CLI keep only the `__version__` test.
1. **Git & Validation**:
   - Run `git init --initial-branch=main`.
   - Run verification sequence (`install` already runs `uv sync`):
     ```bash
     mise run install
     mise run format
     mise run check
     mise run test
     ```
     On a fresh repo, `check:leaks` prints a benign `no commits yet` and exits 0 — nothing to scan until the first commit below.
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
1. **Layout**:
   - `app/agent.py` — defines the `root_agent` symbol and its tools. Tools are plain typed functions; ADK infers each JSON schema from the signature and docstring. Keep business logic in the library/modules and call into it from tools.
   - `app/fast_api_app.py` — FastAPI backend server for API interaction.
1. **Models & Auth**:
   - Pin the current Flash generation by default (e.g. `gemini-3.6-flash`, matching `agents-cli`'s own scaffold) rather than the `-latest` alias — Vertex AI's `-latest` resolution has documented version-ambiguity and hot-swap quality regressions. Check `agents-cli create --agent adk`'s generated default or the [Gemini models list](https://ai.google.dev/gemini-api/docs/models) and bump to the newest Flash generation when one ships; only pin an exact dated snapshot when you need strict reproducibility.
   - Use Google Application Default Credentials (ADC) for authentication. In local development, run `gcloud auth application-default login`.
   - In `.env`, ensure `GOOGLE_GENAI_USE_VERTEXAI=true`, `GOOGLE_CLOUD_PROJECT=<project_id>`, and `GOOGLE_CLOUD_LOCATION=<region>` (e.g., `europe-west1` or `global`) are set.
1. **Development Commands**:
   - Setup project: `uvx google-agents-cli setup`
   - Install dependencies: `agents-cli install`
   - Start local playground with live reload: `agents-cli playground`
   - Run tests: `uv run pytest tests/unit tests/integration`
   - Evaluate agent: `agents-cli eval generate` followed by `agents-cli eval grade`
   - Deploy agent: `agents-cli deploy`
1. **Testing**: Import `root_agent` and assert its wiring (name, model, tools), and exercise tool functions directly — no API key, no mocks, and no web `conftest.py`/Postgres.

## Gotchas & Guidelines

- **`uv init` Python Pin**: `uv init` writes a `.python-version` for the interpreter it resolves, which can be older than `requires-python` and breaks `uv sync`. Run `uv python pin <major.minor>` right after bootstrapping.
- **`Slug` vs `Package` in Import Paths**: When `Slug` contains a hyphen (e.g. `my-cool-app`), substituting it literally into a Python import position — `[project.scripts]`'s right-hand side, `pdoc`/`granian` module arguments, `from <slug> import ...` — produces invalid syntax; `validate-pyproject` (`check:format`) rejects the resulting `pyproject.toml` outright. Use `Package` (underscores) there; `Slug` (hyphens allowed) stays correct for the distribution name, directory arg to `uv init`, Docker tag, and the console-script command itself.
- **`uv_build` Upper Bound**: Keep `[build-system].requires`'s `uv_build` upper bound at least one minor ahead of the pinned `uv` tool version — `uv sync`/`uv build` warns (and a stricter resolver would fail) when the installed `uv_build` version falls outside the declared range.
- **`ty` Python Version**: `[tool.ty.environment].python-version` requires `major.minor` format (e.g., `"3.14"`). Do not supply patch versions.
- **Line Length**: Ruff default line length is 120 characters.
- **`ty` & SQLAlchemyDTO**: The scaffold ships with no blanket `[tool.ty.rules]` ignores and type-checks clean. `ty` is pre-1.0; if it later flags the generated DTO type forms once you add ORM models, add a single scoped `invalid-type-form = "ignore"` rather than a broad `unresolved-*` ignore that would hide real typos and missing imports.
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
