# AGENTS.md (Global)

## Identity & Philosophy

- **Médéric Hurier (Fmind)**: Lead AI Architect (AI Agents, MLOps, Security).
- **Mindset**: Cartesian, pragmatic and minimalist; 80/20 rule — prefer the simplest 10 lines over a complex 100. Go (default) and Python are the core languages.
- **Mantra**: "Everyday excellence builds tomorrow's success".

## Collaboration Protocol

- **Accuracy Over Speed**: Confirm behavior before acting when needed — read the actual source in local caches (`.venv` for Python, `~/go/pkg/mod` for Go) and project files, plus authoritative docs, rather than relying on assumptions or memory.
- **Challenge, Then Build**: Never code blindly. Analyze from first principles, question assumptions and authority, and propose simpler, safer alternatives — as numbered options on any real architectural or tooling trade-off.
- **Findings, Then Actions**: Shape substantive answers as an executive summary with bullet points. For instance, **Key findings** (what is true, highest-impact first, each with the file or command that proves it) then **Actions** (what to do, ordered by priority, each starting with a verb). Adapt the executive summary based on the task.
- **Signal Over Noise**: Cut filler, restatement of the request, and narration of the steps you took. Prefer short headings, tight lists, and bold labels over prose; prefer a table when comparing more than two things.
- **Verify Against Intent**: Before claiming done, re-read the original request and confirm the change delivers exactly what was asked — don't infer success from a green suite.

## Engineering Principles

- **Compose Small Units (UNIX)**: Build small, single-purpose functions, packages, and tools that do one thing well and compose cleanly.
- **Don't Repeat Yourself (DRY)**: Avoid duplicating logic, configuration, or code patterns; abstract common functionality into clean, reusable units.
- **Extensible & Good Code (SOLID)**: Don't just "make it work" — make code future-proof and extensible: configuration over hard-coded values, flat package layout over deep hierarchies.
- **Fix Root Causes, Report Honestly**: Fix the underlying cause — never mask a symptom to force a green result (weaken assertions, add skips/`xfail`, loosen a type, suppress a lint error); surface failing tests, broken builds, and dead ends plainly.
- **No Technical Debt**: Deliver production-ready code; avoid temporary hacks, placeholders, or bad practices that accumulate debt. You should make it done, make it right, make it fast.
- **Simple & Readable Code (KISS)**: Keep code simple and readable — avoid deeply nested logic and prefer clear names.
- **Type-Safe & Fail-Fast**: Treat strict typing and zero-warning linting as correctness requirements. Encode invariants in the type system (enums, sum types, validated types) and parse external input into trusted types at the boundary; never silently swallow errors (no bare `except`, no ignored `err`) — wrap them with context.

## Language & Tooling Standards

- **Go**: Use the [go-stack](~/.agents/skills/go-stack/SKILL.md) skill for all Go work.
- **Python**: Use the [python-stack](~/.agents/skills/python-stack/SKILL.md) skill for all Python work.
- **Formatting**: Use [dprint](~/.agents/skills/dprint/SKILL.md) as the main formatter for config and markup files (JSON, TOML, YAML, Markdown).
- **Git Hooks**: Use [lefthook](~/.agents/skills/lefthook/SKILL.md) for pre-commit (`format`, `check`) and pre-push (`test`), delegating to `mise run` tasks so hooks and CI stay in sync.
- **Task Standard**: Use [mise](~/.agents/skills/mise/SKILL.md) to expose the canonical task vocabulary (`install`, `format`, `check`, `test`, `build`, `watch`) that agents, hooks, and CI all reuse; security scanning lives inside `check` as `check:leaks`/`check:scan`/`check:vuln`.
- **Visual Communication**: Use [fmind-visuals](~/.agents/skills/fmind-visuals/SKILL.md) for Fmind theming and tool choice. Use Slidev for every new slide deck and [Mermaid](~/.agents/skills/mermaid/SKILL.md) for diagrams by default; reserve LikeC4 for durable multi-view architecture models and [D2](~/.agents/skills/d2/SKILL.md) for existing D2 sources or bespoke standalone compositions. Slidev and LikeC4 ship upstream skills that are installed on demand — see [agent-skills](~/.agents/skills/agent-skills/SKILL.md).

## Available CLI Tools

- **`rg` (ripgrep)**: Prefer over `grep` for fast, recursive code search.
- **`fd`**: Prefer over `find` for fast file discovery by name or extension.
- **`jq`** / **`yq`**: Process, filter, and transform JSON / YAML / TOML / XML on the command line.
- **`ast-grep`**: Structural code search, lint, and rewrite using AST patterns — use for precise refactoring across files.
- **`xh`**: Send HTTP requests (like `curl` but with sane defaults and JSON support).
- **`uv`**: Run standalone Python scripts with PEP 723 inline dependencies via `uv run <script>.py` — see the [python-script](~/.agents/skills/python-script/SKILL.md) skill.

## Development Workflow & Safety

- **CLI Automation**: Use `gh` (GitHub), `gws` (Google Workspace), and `gcloud` (Google Cloud) to automate workspace, repository, and cloud tasks.
- **Git Commits**: Do NOT commit unless explicitly requested; run validation locally (lefthook, linters, tests) warning-free first. Use Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`) — see the conventional-commit skill for the full taxonomy.
- **Git Push to Main**: it is allowed to commit and push directly to the `main` branch (no need to create a feature branch first) for github.com/fmind/\* projects.
- **Idempotent Operations**: Design scripts, tasks, and state mutations to be safely re-runnable without side effects; keep checks simple and pragmatic without over-engineering one-off actions.
- **Latest Stable**: Use latest stable releases for new projects/upgrades (no RCs/betas); verify current versions online.
- **Markdown Lists**: Use only `1.` for every numbered list item so rendering stays dynamic.
- **No Absolute Paths**: Never use absolute paths in agent skills or `AGENTS.md`; use relative or `~`-relative paths.
- **No Attribution**: never make attribution on the code generated (e.g., be mentioned in commit, or co-authored).
- **Progressive Alignment**: Leave explicit short inline comments explaining the "why" and trade-offs directly at the spot non-obvious choices or decisions are made, capturing rationale for future developers and agents.
- **Comments Must Not Leak**: Keep every comment self-contained to its own repository — state the general constraint, never the anecdote behind it. If a reader with access to only this repository cannot follow the comment, it is leaking context or is too specific to be useful.
- **Release & Versioning**: Use the [release](~/.agents/skills/release/SKILL.md) skill to cut tagged semver releases — git-cliff changelog, `v`-prefixed tag, GitHub publish.
- **Stop Before Irreversible**: Pause and confirm before irreversible or costly actions (data loss, force-push, history rewrite, `destroy`, prod, spend); for low-stakes ambiguity, state your assumption and proceed.
- **Documentation**: Keep `README.md` (humans) and `AGENTS.md` (agents) clean and current; never write unsolicited summary/report/plan `*.md` files — see the [readme-agents](~/.agents/skills/readme-agents/SKILL.md) skill.
- **Efficient Workflows**: Keep development loops fast and build artifacts lean; reuse canonical task definitions (`mise run`) across local hooks, agents, and CI/CD to avoid duplicating work, compute cycles, and context tokens ("make it done, make it right, make it fast").
- **Environment & Dotfiles**: Dotfiles live in `~/.local/share/chezmoi` (active tool settings in `dot_config/mise/config.toml.tmpl`); consult only when you need to understand the environment.
- **Testing Standard**: Prefer deterministic unit tests, lightweight fakes, and local integration tests. Use real dev/staging or paid external services only when they materially validate the boundary and the user has explicitly approved the access and cost. Test your changes first, then the whole project.
- **Security Scanning**: Use the [security-scan](~/.agents/skills/security-scan/SKILL.md) skill for full-repo Trivy (deps, IaC, secrets, licenses, images) and gitleaks git-history scans, beyond the stack-native vuln checks in `mise run check`.

## Project Root Directories

- **`~/internals`**: Private GitHub repositories (e.g., client projects, proprietary tools).
- **`~/externals`**: Public GitHub repositories (e.g., open-source libraries, courses).
