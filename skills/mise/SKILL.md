---
name: mise
description: Configure pinned mise tools and the canonical install, format, check, test, build, and watch tasks shared by hooks and CI. Use for any mise.toml or task work.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/mise
  created: "2026-07-04"
  updated: "2026-09-03"
---

# Mise

One `mise.toml` per project pins the toolchain and defines the task vocabulary that hooks and CI reuse; mise owns _what_ every task does, while [lefthook](../lefthook/SKILL.md) and [github-actions](../github-actions/SKILL.md) only decide _when_ to run it.

## Task Vocabulary

Every project exposes the same core tasks with short aliases so agents, hooks, and CI stay portable:

| Task      | Alias | Purpose                                                                                                                                                                                 |
| --------- | ----- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `install` | `i`   | Sync dependencies and install git hooks (`lefthook install`).                                                                                                                           |
| `format`  | `f`   | Format all sources (fans out to `format:*`).                                                                                                                                            |
| `check`   | `c`   | All static checks in parallel (fans out to `check:*`).                                                                                                                                  |
| `test`    | `t`   | Run the test suite.                                                                                                                                                                     |
| `build`   | `b`   | Compile or package artifacts (fans out to `build:*`).                                                                                                                                   |
| `watch`   | `w`   | Run the app with live reload, or re-run tests where there is no app to serve; omitted only by a stack with neither, such as [terraform-stack](../terraform-stack/references/mise.toml). |
| `all`     | `a`   | `format`, `check`, `test`, `build` in sequence: the full gate.                                                                                                                          |

Language stacks ship concrete files: [go-stack](../go-stack/references/mise.toml), [python-stack](../python-stack/references/mise.toml), and [angular](../angular/references/mise.toml) for TypeScript web applications.

## Subtask Naming

Split a task into `<task>:<x>` when one piece must run alone; each family keys `<x>` off a different noun:

- **`format:<input>`**: the source family formatted — `format:go`, `format:python`, `format:templ`; `format:dprint` for JSON, Markdown, TOML, and YAML.
- **`build:<output>`**: the artifact produced — `build:go` (binary), `build:templ` (Templ to Go), `build:css`, `build:js`, `build:docs`, `build:image` (OCI image).
- **`check:<concern>`**: the property verified, identical across languages so `mise run check:lint` means the same everywhere; the names are fixed:

| Task            | Concern                          | Tool                                                            |
| --------------- | -------------------------------- | --------------------------------------------------------------- |
| `check:format`  | formatting drift                 | `dprint check` plus the stack formatter's check mode            |
| `check:lint`    | lint rules                       | `golangci-lint`, `ruff`, Biome                                  |
| `check:types`   | static types                     | `ty`, `tsc --noEmit`                                            |
| `check:vuln`    | dependency CVEs                  | `govulncheck`, `uv audit`, `pnpm audit`                         |
| `check:leaks`   | committed secrets                | [gitleaks](../gitleaks/SKILL.md)                                |
| `check:scan`    | IaC and config misconfigurations | [trivy](../trivy/SKILL.md)                                      |
| `check:actions` | workflow lint and audit          | `actionlint` + [zizmor](../zizmor/SKILL.md)                     |
| `check:sast`    | insecure code patterns (opt-in)  | [opengrep](../opengrep/SKILL.md), only when a project adopts it |

Those names are reserved: never respell one (`check:audit`, `check:dprint`) when the table already covers the concern. A stack adds a name only for a concern the table has none for, and the shipped set is closed: `check:deps` (unused files and dependencies), `check:doc` (document compiles), `check:pkg` (publishable surface), `check:site` (site builds clean), `check:validate` (configuration syntax). In a polyglot repository such as the dot repository root, a concern that repeats per language may split by language (`check:go`, `check:python`, `check:shell`), while a genuinely shared concern keeps its cross-language name (`check:format` for the one dprint check). Aliases are best-effort: a repository that already spends `f`, `t`, or `i` keeps them; the task names are the contract.

## Conventions

- **Hooks**: see [lefthook](../lefthook/SKILL.md); each hook command is `mise run <task>` and its name mirrors the task.
- **Parallel checks**: `check` fans out with `depends = ["check:format", "check:lint", "check:types", "check:vuln"]`; mise runs the subtasks concurrently.
- **Incremental tasks**: declare `sources` and `outputs` so mise skips a task whose inputs are unchanged (ideal for builds).
- **Staged vs whole-tree**: only formatters take `{staged_files}`; `check` and `test` always run on the whole tree.
- **Argument passthrough**: mise appends CLI args to the task's last command; `$@` is empty in TOML tasks and `{{arg()}}` is deprecated, so use the `usage` field when args must land elsewhere.
- **Dotenv**: `[env]` with `_.file = ".env"` auto-loads the file.

## Tool Management

```bash
mise registry <name>     # discover the tool's backend id
mise use <tool>@latest   # pin into [tools] and install
mise install             # install everything pinned
mise lock                # refresh mise.lock for reproducibility
mise upgrade --bump      # bump pinned versions (updates mise.lock too)
```

## Gotchas

- **Full gate on a dirty tree**: `mise run all` write-formats the whole tree; when unrelated changes are present, run it in a temporary `git worktree` or fall back to `mise run check` and `mise run test`.
- **Trust**: in normal mode `mise run`, `mise install`, `mise exec`, and `mise watch` trust the active config automatically; `mise trust` is only needed for other commands or in paranoid mode.
- **Fail fast in hooks**: set `run_auto_install = false` under `[settings.task]` so a missing tool errors instead of installing silently.
- **Non-interactive scripts**: pass `-y` (`mise install -y`) in scripts and CI steps that would otherwise prompt.
- **Keep project config project-local**: never symlink a repository's `mise.toml` into `~/.config/mise/conf.d/`; mise then treats it as global, `mise lock` reports `No tools configured to lock`, and its tasks leak everywhere.
- **Task `dir`**: defaults to the config root (the repo locally, the checkout in CI), so no `[task_config] dir` override is needed.

## Documentation

- [mise](https://mise.jdx.dev) · [Tasks](https://mise.jdx.dev/tasks/) · [Settings](https://mise.jdx.dev/configuration/settings.html)
- Companion skills: [lefthook](../lefthook/SKILL.md) (hooks call these tasks), [github-actions](../github-actions/SKILL.md) (CI installs the toolchain with `mise-action` and runs `mise run all`).
