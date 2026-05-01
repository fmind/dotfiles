---
name: use-gemini-cli
description: Guide for using the gemini CLI — interactive sessions, headless / scripted mode (`-p` + JSONL), key flags, in-session slash commands, sessions, worktrees, and CI exit codes.
---

# Use Gemini CLI

`gemini` is the Gemini CLI binary. It runs in two modes: **interactive** (TTY, default) and **headless** (non-TTY or `-p`, JSONL events).

Configuration is layered (system → user → workspace) — this skill is about *invocation*; for config schema see `configure-gemini-cli`.

## One-time Setup

```bash
# Install (already pinned via mise in this dotfiles repo).
mise install gemini

# First run — pick an auth method (OAuth | Gemini API key | Vertex AI ADC | SA | Vertex API key).
gemini
# /auth                                # switch auth method later (no `gemini auth` subcommand exists)

# Trust the current folder once per repo, per machine (one-shot bypass: --skip-trust).
```

## Interactive Sessions

```bash
# Start in cwd.
gemini

# Launch with a model override and approval mode preset.
# Check `gemini /model` (in-session) or the docs for current model IDs.
gemini -m <model-id> --approval-mode auto_edit

# Resume the most recent session (or pick interactively).
gemini -r                               # most recent
gemini -r latest                        # explicit "latest"
gemini -r 5                             # by index (see --list-sessions)
gemini --list-sessions
gemini --delete-session 5               # by index

# Launch with an isolated git worktree (experimental.worktrees must be on).
gemini -w feat/new-thing                # creates worktree branch + working dir
gemini -w                               # auto-generates a name (worktree-<hash>)

# Pull extra dirs into the workspace context.
gemini --include-directories ../shared,../proto

# Restrict the MCP surface (handy when reviewing untrusted code).
# Tool-level restriction is now the Policy Engine, not --allowed-tools (deprecated).
gemini --allowed-mcp-server-names github \
       --policy ./.gemini/policies/review.toml
```

## Headless / Scripted Mode

When stdout is not a TTY, or `-p`/`--prompt` is passed, Gemini emits a single response (or JSONL stream) and exits. This is the form used in CI and shell pipelines.

```bash
# One-shot prompt, plain text out.
gemini -p "summarize CHANGELOG.md"

# Stream JSONL events (init | message | tool_use | tool_result | error | result).
gemini -p "list TODO comments under src/" -o stream-json

# Single JSON object out (good for piping to jq).
gemini -p "what changed since v1.4.0?" -o json | jq -r '.result'

# Read prompt from stdin (works because stdin is not a TTY).
git diff --cached | gemini -p "write a conventional commit message"

# Interactive prompt with pre-seeded text (still TTY).
gemini -i "draft a PR description for the staged changes"

# YOLO mode (auto-approve every tool call) — use only in throwaway sandboxes.
gemini -p "scaffold a uv project here" -y
```

### JSONL event types (`-o stream-json`)

| Event | When |
|-------|------|
| `init` | Session started; carries session id, cwd, model |
| `message` | User / assistant message chunk |
| `tool_use` | Tool call about to run (name + args) |
| `tool_result` | Tool finished (truncated payload) |
| `error` | Non-fatal warning during the run |
| `result` | Final answer + token / cost usage |

### Exit codes

| Code | Meaning |
|------|---------|
| `0`  | Success |
| `1`  | Generic / API error (auth, network, runtime) |
| `42` | Input validation failure (bad CLI args / unknown flag) |
| `53` | Hit `model.maxSessionTurns` |

## Common Flags

| Flag | Purpose |
|------|---------|
| `-m, --model <id>` | Override model for this session — see `/model` or the [model reference](https://geminicli.com/docs/reference/configuration/) for current IDs |
| `-p, --prompt <text>` | Headless one-shot prompt (appended to stdin if any) |
| `-i, --prompt-interactive <text>` | TTY session pre-seeded with text |
| `-y, --yolo` | Auto-approve every tool call |
| `--approval-mode <m>` | `default` / `auto_edit` / `yolo` / `plan` |
| `-s, --sandbox` | Run tools inside the configured sandbox |
| `-r, --resume [latest\|<index>]` | Resume a prior session |
| `--list-sessions`, `--delete-session <index>` | Session housekeeping |
| `-w, --worktree [name]` | Run inside an isolated git worktree (auto-named if omitted) |
| `-e, --extensions <a,b>` | Limit active extensions for this run |
| `-l, --list-extensions` | Show installed extensions |
| `--policy <paths>` | Extra Policy Engine files / dirs to load |
| `--admin-policy <paths>` | Admin-tier policy files (override workspace + user) |
| `--allowed-mcp-server-names <a,b>` | MCP server allowlist |
| `--allowed-tools <a,b>` | **Deprecated** — use the Policy Engine instead |
| `--include-directories <a,b>` | Extra workspace roots |
| `-o, --output-format <fmt>` | `text` (default) / `json` / `stream-json` |
| `--raw-output` | Disable model-output sanitization (allows ANSI escapes — risky on untrusted output) |
| `--accept-raw-output-risk` | Suppress the `--raw-output` security warning |
| `--skip-trust` | One-shot trust bypass (CI) |
| `--acp` | Speak Agent Communication Protocol (IDE bridge); `--experimental-acp` is deprecated |
| `--screen-reader` | Plain-text UI for screen readers |
| `-d, --debug` | Verbose logs (F12 opens debug console in interactive mode) |

## Subcommands (outside a session)

```bash
# MCP servers — registry-style management.
gemini mcp list                                # connection state of every MCP server
gemini mcp add <name> <commandOrUrl> [args...] # register an MCP server
gemini mcp remove <name>
gemini mcp enable <name> | disable <name>

# Extensions — bundles of MCP servers, commands, agents, skills.
gemini extensions list                         # see also: configure-gemini-cli-extensions
gemini extensions install <git-url-or-path> [--auto-update] [--pre-release]
gemini extensions update [<name>] [--all]
gemini extensions uninstall <name..>
gemini extensions enable|disable [--scope] <name>
gemini extensions link <path>                  # live-link a local extension
gemini extensions new <path> [template]        # scaffold a new extension
gemini extensions validate <path>

# Agent skills.
gemini skills list [--all]
gemini skills install <git-url-or-path> [--scope] [--path]
gemini skills enable|disable <name> [--scope]
gemini skills link <path>
gemini skills uninstall <name> [--scope]

# Hooks (migration helper for now).
gemini hooks migrate                           # import Claude Code hooks

# Local Gemma routing (run small models on-device via LiteRT-LM).
gemini gemma setup                             # download + configure local routing
gemini gemma start | stop | status | logs
```

## In-session Slash Commands

```text
# Identity, auth, model
/auth                  Switch auth method (OAuth, API key, Vertex)
/model                 Show / set the active model (manage, set)
/about                 Version + build info (paste this when filing issues)
/privacy               Privacy notice and consent options

# Context & memory
/init                  Generate a tailored GEMINI.md for the cwd
/memory                add | list | show | refresh | inbox  (see configure-gemini-cli-memory)
/directory             Manage workspace directories (add, show)
/compress              Replace chat history with a summary to save tokens
/rewind                Step backward through conversation history

# Tools, MCP, extensions, skills, hooks, policies
/tools                 Currently enabled core tools
/mcp                   list / desc <server> / schema <server>
/extensions            List, enable, disable extensions for this session
/skills                List, enable, disable, reload agent skills
/hooks                 Manage lifecycle event hooks
/agents                List / enable / disable / reload subagents
/commands              Manage custom slash commands (list, reload)
/permissions           Workspace trust + shell allow/deny
/policies              Show effective policy bundle

# Modes & sessions
/plan                  Toggle plan mode (read-only; routes between reasoning and fast tiers)
/restore               Roll back to a prior checkpoint (shadow git)
/resume                Browse / resume prior sessions (alias: /chat)
/stats                 Session, model, tool usage stats
/shells                Toggle background shells view

# UI / IDE / setup
/ide                   Toggle IDE companion (VS Code / Antigravity / JetBrains via ACP)
/theme                 Switch theme
/editor                Pick external text editor
/vim                   Toggle vim mode for input
/terminal-setup        Configure terminal keybindings
/settings              Open the settings editor
/setup-github          Scaffold the Gemini Code Assist GitHub App

# Help & meta
/help                  Show in-session help
/docs                  Open documentation in browser
/upgrade               Open the upgrade page
/bug                   File a GitHub issue about Gemini CLI
/copy                  Copy last output to clipboard
/clear                 Clear screen and scrollback
/quit                  Exit (supports `--delete` to wipe the session on the way out)
```

## Worktree Workflow

With `experimental.worktrees: true` in `settings.json`:

```bash
gemini -w feat/auth-rewrite             # creates worktree at .gemini/worktrees/feat/auth-rewrite
gemini -w                               # auto-named worktree (e.g., worktree-a1b2c3d4)
# Hack inside session...
/quit                                   # leaves worktree + branch intact (preserves work)
git worktree list                       # confirm
git worktree remove .gemini/worktrees/feat/auth-rewrite --force
git branch -D feat/auth-rewrite
```

Two parallel `gemini -w` runs against the same repo do not collide — each owns its working tree. Cleanup is intentionally manual; Gemini never deletes a worktree on exit.

## Common Workflows

**Run a quick analysis without polluting the repo.**

```bash
gemini -p "list every function over 60 lines in src/" -o json | jq -r '.result'
```

**Use Gemini in a pre-commit hook.**

```bash
git diff --cached | gemini -p "review this diff; output BLOCK or PASS only" \
  --approval-mode default --skip-trust
```

**Chain headless calls in a pipeline.**

```bash
gemini -p "extract TODO comments as JSON" -o json \
  | jq -r '.result' \
  | gemini -p "group by file, return markdown table"
```

## Important Notes

1. **Pick the right approval mode for the context.** `yolo` in CI on someone else's repo will execute arbitrary tool calls — use `default` + `--skip-trust` when running on untrusted input.
2. **Headless mode auto-detects a non-TTY**, so piping in/out usually does the right thing without `-p`. Use `-o stream-json` when the consumer needs progress events.
3. **Exit code `53`** means the session ran out of turns — bump `model.maxSessionTurns` (or set `-1`) before re-running long jobs.
4. **`--allowed-mcp-server-names` is a per-invocation allowlist** layered on top of `settings.json` — it can only narrow, never widen. **Tool-level restriction is now Policy Engine territory** (`--policy`, `--admin-policy`, plus TOML rules under `~/.gemini/policies/`); `--allowed-tools` is deprecated and may be removed.
5. **Worktrees keep parallel sessions safe**, but they share the same `~/.gemini/` user state (sessions, GEMINI.md, skills) — workspace state is per-worktree under `.gemini/worktrees/`.
6. **`--skip-trust` should be reserved for CI / one-off runs on known code.** It bypasses the folder-trust prompt and re-enables project settings, MCP servers, hooks, and extensions in one go.
7. **`--raw-output` disables sanitization of model output** (e.g., ANSI escapes pass through). Only use it on trusted output, and pair with `--accept-raw-output-risk` to silence the warning in scripts.

## Documentation

- [Gemini CLI docs (root)](https://geminicli.com/docs)
- [Headless / scripting mode](https://geminicli.com/docs/cli/headless/)
- [Session management](https://geminicli.com/docs/cli/tutorials/session-management/)
- [Git worktrees (experimental)](https://geminicli.com/docs/cli/git-worktrees/)
- [Slash commands reference](https://geminicli.com/docs/reference/commands/)
- [Policy Engine](https://geminicli.com/docs/reference/policy-engine)
- Companion skills: `configure-gemini-cli`, `configure-gemini-cli-hooks`, `configure-gemini-cli-memory`, `configure-gemini-cli-extensions`, `setup-gemini-cli-on-new-project`.
