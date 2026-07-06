---
name: mise
description: Canonical mise setup — the per-project task vocabulary (install, format, check, test, build, watch) and pinned toolchain that hooks and CI reuse. Use when defining `mise.toml` tasks or pinning tools.
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/mise
  created: 2026-07-04
  updated: 2026-07-06
---

# Mise Standard

Canonical setup for **mise** as the per-project task runner and tool-version manager. One `mise.toml` defines the task vocabulary that git hooks and CI reuse, and pins the toolchain for reproducibility. Mise is the **single source of truth** for what every check does; [lefthook](../lefthook/SKILL.md) and [github-actions](../github-actions/SKILL.md) only decide _when_ to run them.

## Task Vocabulary

Every project exposes the same core tasks (with short aliases) so agents, hooks, and CI stay portable. In interactive shells, `mr <task>` is the standard command-line alias used for `mise run <task>`:

| Task      | Alias | Purpose                                                       |
| --------- | ----- | ------------------------------------------------------------- |
| `install` | `i`   | Sync dependencies and install git hooks (`lefthook install`). |
| `format`  | `f`   | Format all sources (fans out to `format:*`).                  |
| `check`   | `c`   | All static checks in parallel (fans out to `check:*`).        |
| `test`    | `t`   | Run the test suite.                                           |
| `build`   | `b`   | Compile or package artifacts (fans out to `build:*`).         |
| `watch`   | `w`   | Run the app with live reload (web).                           |

Language stacks ship concrete `mise.toml` files — see [go-stack](../go-stack/references/mise.toml) and [python-stack](../python-stack/references/mise.toml).

## Naming: generic vs specific

Split a task into `<task>:<x>` subtasks when you want to run one piece on its own. Each family keys `<x>` off a **different noun**, because each answers a different question:

- **`format:<input>`** — key off the _source you format_ (the file family): `format:go`, `format:python`, `format:templ`, `format:shell`. Fall back to the **tool name** only when the domain has no single language noun: `format:dprint` (JSON/Markdown/TOML/YAML).
- **`check:<concern>`** — key off the _property verified_, tool-agnostic and **identical across languages** so `mise run check:lint` means the same everywhere: `check:format`, `check:lint`, `check:types`, `check:vuln`, `check:leaks` (secret/credential scan, e.g. `gitleaks`), `check:scan` (IaC/config misconfig, e.g. `trivy config`). Use one name per concern — dependency-vulnerability scanning is always `check:vuln` (never `check:audit`), whether it runs `govulncheck` or `pip-audit`; secret scanning is always `check:leaks`; misconfiguration scanning is `check:scan`, distinct from `check:vuln`.
- **`build:<output>`** — key off the _artifact produced_: `build:go` (binary), `build:css`, `build:html` (generated Templ→Go).

So `format:templ` (you format `.templ` source) and `build:html` (you generate the HTML-rendering Go) coexist for the same tech — the noun differs because the question differs. In a **polyglot repo** where a concern repeats across languages (a dotfiles root linting Go + Python + Shell), subdivide _that_ concern by **language** (`check:go`, `check:python`, `check:shell`), each bundling its language's concerns — but a genuinely shared concern keeps its cross-language name (`check:format` for the one dprint config/markup check).

**Aliases are best-effort:** the single-letter aliases above are the target, but a management-heavy repo that already spends `f`/`t`/`i` on daily tasks keeps those and lets the canonical tasks use their full names — the task _names_ are the contract, aliases are convenience.

## Conventions

- **Delegate hooks to tasks**: pre-commit/pre-push call `mise run <task>` so local hooks and CI stay identical — see the [lefthook skill](../lefthook/SKILL.md). Command names mirror the task they call (`format:go`, `check`).
- **Parallel checks**: fan `check` out with `depends = ["check:format", "check:lint", "check:types", "check:vuln"]` and namespaced subtasks — mise runs them concurrently.
- **Incremental tasks (caching)**: Define `sources` and `outputs` arrays to cache execution. When input files under `sources` are unmodified, `mise` skips running the task and uses the cached `outputs` (ideal for builds).
- **Staged vs whole-tree**: pass `{staged_files}` only to **formatters** (fast, and they restage their fixes); `check`/`test` always run the whole tree so correctness stays global.
- **Argument passthrough (auto-append, not `$@`)**: mise appends any CLI args to the **end of the task's last command** and does **not** populate shell `$@`/`${@}` in TOML tasks — those expand empty, so `${@:-x}` silently collapses to the literal `x` and its "passthrough" never fires. Rely on the auto-append and pick the lightest form that stays correct:
  1. **bare** when the tool already defaults sensibly with no path (`dprint fmt`, `ruff check`, `pytest`) — auto-append then also lets a formatter honor its `{staged_files}` for free.
  1. **hardcoded default** when the tool needs an explicit target (`govulncheck ./...`, `gofumpt -w .`, `go test … ./...`); extra args append _on top_ of it (additive, not replace).
  1. **delete** a bare `${@}` — it is a pure no-op.

  Auto-append reaches only the **last** command. Avoid the `{{arg()}}`/`{{option()}}`/`{{flag()}}` Tera helpers (deprecated); if you genuinely need default-_and_-replace, or args in more than one command, use the `usage` field.

- **Dotenv**: auto-load env with `[env]` `_.source = ".env"`.

## Tool Management

Pin the toolchain in `[tools]` and resolve versions from the registry:

```bash
mise registry <name>     # discover the tool's backend id
mise use <tool>@latest   # pin into [tools] and install
mise install             # install everything pinned
mise lock                # refresh mise.lock for reproducibility
mise upgrade --bump      # bump pinned versions, then re-lock
```

## Gotchas

- **Trust**: `mise` requires `mise trust` before running a new project's config.
- **Fail fast in hooks**: set `run_auto_install = false` under `[settings.task]` so hooks error on a missing tool instead of silently installing it.
- **Non-interactive execution**: Always pass `-y`/`--yes` to `mise` commands in automated scripts or workflows (e.g., `mise trust -y`, `mise install -y`) to prevent blocking on prompts.
- **Portable task dirs**: if the same `mise.toml` is exposed globally (e.g. via a `conf.d` symlink), a hardcoded `[task_config] dir` breaks in CI where the source isn't deployed — fall back to `$GITHUB_WORKSPACE`, e.g. `dir = "{{ env.GITHUB_WORKSPACE | default(value='~/path') }}"`.

## Documentation

- [mise Documentation](https://mise.jdx.dev)
- [github-actions](../github-actions/SKILL.md) — CI installs this toolchain via `mise-action` and reuses these tasks.
