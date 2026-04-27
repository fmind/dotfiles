---
name: configure-gemini-cli-memory
description: Guide for Gemini CLI's memory & context system — `GEMINI.md` / `AGENTS.md` resolution order, modular `@file.md` imports, `.geminiignore`, `context.fileFiltering` settings, `experimental.autoMemory` + `/memory inbox`, and the `save_memory` tool.
---

# Configure Gemini CLI Memory

Gemini CLI's "memory" is the bundle of markdown that gets prepended to every prompt. It is assembled from multiple files at session start, refreshed on demand via `/memory refresh`, and can be extended at runtime by the `save_memory` tool or Auto Memory promotion.

This skill covers the static context layer (`GEMINI.md` / `AGENTS.md`, `.geminiignore`) and the dynamic layer (Auto Memory, `save_memory`, `/memory ...` slash commands). For session-level facts that should travel across sessions, prefer this layer over re-pasting context.

## Resolution Order

When a session starts, Gemini CLI loads context files in this order and **concatenates** them:

1. **Global** — `~/.gemini/GEMINI.md` (always first; includes the `## Gemini Added Memories` section maintained by `save_memory`).
2. **Workspace + ancestors** — every `GEMINI.md` (or whatever names you list in `context.fileName`) found from the cwd upward to the trusted root.
3. **Just-in-time** — when a tool touches a path, the CLI re-scans that path and its ancestors so subdirectory `GEMINI.md` files get pulled in only when relevant (`experimental.jitContext`).

There is **no supersedence** between layers — they all stack. Order matters only for who-overrides-whom in case of contradicting instructions (later wins).

## `context.fileName`

```json
{
  "context": {
    "fileName": ["AGENTS.md", "GEMINI.md"]
  }
}
```

- Accepts a string or an array.
- Default: `"GEMINI.md"`.
- When multiple names are listed, **all matching files load and concatenate** — `AGENTS.md` does not replace `GEMINI.md`. Use both if you want a single repo to feed Gemini CLI, Claude Code, and Cursor from one source.

## Modular `@file.md` Imports

Inside any context file, embed another markdown file with `@<path>`:

```markdown
# Project Conventions

@./conventions/style.md
@./conventions/security.md
@../shared/glossary.md
```

- Both relative (`@./...`, `@../...`) and absolute paths are supported.
- Imports happen at memory-load time, not inside the model — the result is one big concatenated buffer.
- Cycle detection behavior is unspecified; keep imports tree-shaped.

Use this to keep `GEMINI.md` short at the top level and split deep guidance into sibling files.

## `.geminiignore` vs `.gitignore`

`.geminiignore` is a sibling to `.gitignore` that controls **what Gemini CLI's tools and context-loader can see**. Both are honored by default:

```json
{
  "context": {
    "fileFiltering": {
      "respectGitIgnore":     true,
      "respectGeminiIgnore":  true,
      "enableRecursiveFileSearch": true,
      "enableFuzzySearch":    true,
      "customIgnoreFilePaths": []
    }
  }
}
```

| Setting | Default | Notes |
|---------|---------|-------|
| `respectGitIgnore` | `true` | Skip files matching `.gitignore` |
| `respectGeminiIgnore` | `true` | Skip files matching `.geminiignore` |
| `enableRecursiveFileSearch` | `true` | Allow `glob`/`grep` to recurse |
| `enableFuzzySearch` | `true` | Fuzzy matching in file pickers |
| `customIgnoreFilePaths` | `[]` | Extra ignore files; **earlier entries take precedence** |

`.geminiignore` is independent of git, so use it to hide files you DO commit but DON'T want the agent to read (large generated fixtures, vendored protos, secret-like sample data).

## `GEMINI.md` vs `AGENTS.md`

Neither supersedes. The recommendation in this dotfile repo (and the user's `~/.gemini/settings.json`) is:

```json
{ "context": { "fileName": ["AGENTS.md", "GEMINI.md"] } }
```

- **`AGENTS.md`** — tool-agnostic project rules read by Gemini CLI, Claude Code, Cursor, OpenCode, and others. Repo-wide conventions belong here.
- **`GEMINI.md`** — Gemini CLI-specific persona / workflow notes. Use it for things only Gemini CLI honors (subagent routing hints, slash-command nudges, model-specific quirks).

Keep `AGENTS.md` first in the array so cross-tool rules load before Gemini-specific overrides.

## Auto Memory (`experimental.autoMemory`)

```json
{ "experimental": { "autoMemory": true } }
```

When enabled, Gemini CLI mines past sessions and proposes **Agent Skills** (multi-step procedures, not single facts) to add to `~/.gemini/skills/` or `<project>/.gemini/skills/`.

- Trigger: ≥10 user messages across recent sessions, **and** sessions have been idle for 3+ hours, **and** lock-coordinated (won't fire mid-session).
- Review: `/memory inbox` opens a dialog with each draft — **promote** (write to skills/), **discard**, or **patch** (edit before promoting).
- Storage: drafts under the project memory dir; transcripts in `~/.gemini/tmp/<project>/chats/`.

Promote sparingly — every promoted skill loads its frontmatter at every session, so a noisy inbox bloats the description budget over time.

## `save_memory` Tool

Available to the model as a built-in tool. One arg: `fact` (a self-contained natural-language statement). Behavior:

- Appends a bullet under `## Gemini Added Memories` in **`~/.gemini/GEMINI.md`** (user-global, NOT workspace).
- Persists across sessions.
- Use for things like "user prefers structlog over stdlib logging" — never for credentials or workspace-specific paths.

## `/memory` Slash Commands

| Command | Effect |
|---------|--------|
| `/memory add <text>` | Append a bullet to `~/.gemini/GEMINI.md` |
| `/memory list` | Paths of every context file currently loaded |
| `/memory show` | Concatenated current memory buffer (debug) |
| `/memory refresh` | Reload from disk without restarting the session |
| `/memory inbox` | Auto Memory promotion dialog (requires `experimental.autoMemory`) |

The separate **`/init`** slash command analyzes the cwd and writes a tailored `GEMINI.md` for the project — handy first step when bootstrapping a new repo (see `setup-gemini-cli-on-new-project`).

## Recommended Project Layout

```text
.
├── AGENTS.md                  # tool-agnostic project rules (cross-agent)
├── GEMINI.md                  # Gemini CLI-specific persona / overrides (optional)
├── .geminiignore              # paths to hide from the agent only
└── .gemini/
    ├── settings.json          # context.fileName, fileFiltering tweaks
    └── skills/                # auto-memory promotion target
```

Workspace memory checked into git → travels with the repo. User memory in `~/.gemini/` stays personal.

## Example `.geminiignore`

```gitignore
# Generated / vendored — agent doesn't need to read these.
**/generated/**
**/__pycache__/**
**/node_modules/**
**/*.lock

# Large fixtures.
tests/fixtures/large/**

# Secrets-shaped sample data we DO commit but don't want sent to the model.
docs/examples/secrets-sample.json
```

## Important Notes

1. **All context files concatenate, including across `context.fileName` entries.** Adding `AGENTS.md` does not replace `GEMINI.md`; both load. Plan for the combined token budget.
2. **`save_memory` writes user-global, not workspace.** It does NOT add to the project `GEMINI.md`. To pin a fact to the repo, write it manually into `AGENTS.md` / `GEMINI.md` and commit.
3. **`.geminiignore` is independent of `.gitignore`.** Use it for committed-but-unhelpful files (large fixtures, generated bundles, vendored deps).
4. **Auto Memory drafts skills, not facts.** Single-fact corrections still need `/memory add` or a `save_memory` call. Review the inbox before promoting — drafts can be wrong.
5. **`@file.md` imports happen at load time.** A broken path silently drops content; check `/memory show` after restructuring.
6. **`/memory refresh` is the fastest debug loop** when iterating on `GEMINI.md` / `AGENTS.md` — no need to restart the session.
7. **JIT context (`experimental.jitContext`) means subdirectory `GEMINI.md` files load lazily.** Don't rely on a deep file being present at session start; expect it to appear after the first relevant tool call.

## Documentation

- [GEMINI.md context files](https://geminicli.com/docs/cli/gemini-md/)
- [.geminiignore](https://geminicli.com/docs/cli/gemini-ignore/)
- [Auto Memory](https://geminicli.com/docs/cli/auto-memory/)
- [`save_memory` tool](https://geminicli.com/docs/tools/memory/)
- [Configuration reference (`context`, `experimental.autoMemory`)](https://geminicli.com/docs/reference/configuration/)
- [Slash commands reference](https://geminicli.com/docs/reference/commands/)
- Companion skills: `configure-gemini-cli`, `use-gemini-cli`, `setup-gemini-cli-on-new-project`, `create-agent-skill`.
