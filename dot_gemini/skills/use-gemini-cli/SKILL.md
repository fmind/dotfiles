---
name: use-gemini-cli
description: Guide for using the gemini CLI — interactive sessions, headless / scripted mode (`-p` + JSONL), key flags, in-session slash commands, sessions, worktrees, and CI exit codes.
---

# Use Gemini CLI

`gemini` is the Gemini CLI binary. It runs in two modes: **interactive** (TTY, default) and **headless** (non-TTY or `-p`, JSONL events). Configuration is layered (system → user → workspace) — this skill is about *invocation*; for config schema see `configure-gemini-cli`.

## One-time Setup

```bash
# Install (already pinned via mise in this dotfiles repo).
mise install gemini

# First run — pick an auth method (OAuth | Gemini API key | Vertex AI ADC | SA | Vertex API key).
gemini
# /auth                                # switch auth method later

# Trust the current folder once per repo, per machine (one-shot bypass: --skip-trust).
```

## Interactive Sessions

```bash
# Start in cwd.
gemini

# Launch with a model override and approval mode preset.
gemini -m gemini-2.5-pro --approval-mode auto_edit

# Resume the most recent session (or pick interactively).
gemini -r
gemini --list-sessions
gemini --delete-session <id>

# Launch with an isolated git worktree (experimental.worktrees must be on).
gemini -w feat/new-thing                # creates worktree branch + working dir

# Pull extra dirs into the workspace context.
gemini --include-directories ../shared,../proto

# Restrict tool surface (handy when reviewing untrusted code).
gemini --allowed-tools read_file,glob,grep_search \
       --allowed-mcp-server-names github
```

## Headless / Scripted Mode

When stdout is not a TTY, or `-p`/`--prompt` is passed, Gemini emits a single response (or JSONL stream) and exits. This is the form used in CI and shell pipelines.

```bash
# One-shot prompt, plain text out.
gemini -p "summarize CHANGELOG.md"

# Stream JSONL events (init | message | tool_use | tool_result | result).
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
| `message` | Model output chunk |
| `tool_use` | Tool call about to run (name + args) |
| `tool_result` | Tool finished (truncated payload) |
| `result` | Final answer + token / cost usage |

### Exit codes

| Code | Meaning |
|------|---------|
| `0`  | Success |
| `1`  | Generic error (auth, network, runtime) |
| `42` | Bad CLI args / unknown flag |
| `53` | Hit `model.maxSessionTurns` |

## Common Flags

| Flag | Purpose |
|------|---------|
| `-m, --model <id>` | Override model for this session |
| `-p, --prompt <text>` | Headless one-shot prompt |
| `-i, --prompt-interactive <text>` | TTY session pre-seeded with text |
| `-y, --yolo` | Auto-approve every tool call |
| `--approval-mode <m>` | `default` / `auto_edit` / `yolo` / `plan` |
| `-s, --sandbox` | Run tools inside the configured sandbox |
| `-r, --resume [id]` | Resume a prior session |
| `--list-sessions`, `--delete-session <id>` | Session housekeeping |
| `-w, --worktree <branch>` | Run inside an isolated git worktree |
| `-e, --extensions <a,b>` | Limit active extensions for this run |
| `-l, --list-extensions` | Show installed extensions |
| `--allowed-tools <a,b>` | Tool allowlist for this run |
| `--allowed-mcp-server-names <a,b>` | MCP server allowlist |
| `--include-directories <a,b>` | Extra workspace roots |
| `-o, --output-format <fmt>` | `text` (default) / `json` / `stream-json` |
| `--skip-trust` | One-shot trust bypass (CI) |
| `--acp` | Speak Agent Communication Protocol (IDE bridge) |
| `--screen-reader` | Plain-text UI for screen readers |
| `-d, --debug` | Verbose logs |

## Subcommands (outside a session)

```bash
gemini auth login                       # interactive auth picker
gemini extensions list                  # see also: configure-gemini-extensions
gemini extensions install <repo>
gemini extensions update --all
gemini mcp list                         # connection state of every MCP server
gemini mcp call <server> <tool> --args '{"k":"v"}'
gemini skills list                      # skills available in this scope
```

## In-session Slash Commands

```text
/auth                  Switch auth method (OAuth, API key, Vertex)
/agents                List / enable / disable / reload subagents
/mcp                   list / desc <server> / schema <server>
/extensions            List, enable, disable extensions for this session
/permissions           Workspace trust + shell allow/deny
/tools                 Currently enabled core tools
/theme                 Switch theme
/ide                   Toggle IDE companion (VS Code / Antigravity / JetBrains via ACP)

/init                  Generate a tailored GEMINI.md for the cwd
/memory add|list|show|refresh|inbox       (see configure-gemini-cli-memory)
/plan                  Toggle plan mode (read-only; routes to Pro then Flash)
/restore               Roll back to a prior checkpoint (shadow git)
/resume                Resume a prior session
/chat                  Save / load named chats

/policies              Show effective policy bundle
/setup-github          Scaffold the Gemini Code Assist GitHub App
```

## Worktree Workflow

With `experimental.worktrees: true` in `settings.json`:

```bash
gemini -w feat/auth-rewrite             # creates worktree at ~/.gemini/worktrees/<repo>-<branch>
# Hack inside session...
exit
git worktree list                       # confirm
git worktree remove ~/.gemini/worktrees/<dir>
```

Two parallel `gemini -w` runs against the same repo do not collide — each owns its working tree.

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
4. **`--allowed-tools` / `--allowed-mcp-server-names` are per-invocation allowlists** layered on top of `settings.json` — they only narrow, never widen.
5. **Worktrees keep parallel sessions safe**, but they share the same `~/.gemini/` user state (sessions, GEMINI.md, skills) — workspace state is per-worktree.
6. **`--skip-trust` should be reserved for CI / one-off runs on known code.** It bypasses the folder-trust prompt and re-enables project settings, MCP servers, hooks, and extensions in one go.

## Documentation

- [Gemini CLI commands & flags](https://geminicli.com/docs/cli/)
- [Headless / scripting mode](https://geminicli.com/docs/cli/headless/)
- [Sessions](https://geminicli.com/docs/cli/sessions/)
- [Worktrees (experimental)](https://geminicli.com/docs/cli/worktrees/)
- [Slash commands reference](https://geminicli.com/docs/reference/commands/)
- [Sandbox & approval modes](https://geminicli.com/docs/cli/sandbox/)
- Companion skills: `configure-gemini-cli`, `configure-gemini-cli-hooks`, `configure-gemini-cli-memory`, `configure-gemini-extensions`, `setup-gemini-cli-on-new-project`.
