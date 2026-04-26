---
name: configure-claude-code
description: Guide for configuring Claude Code via settings.json, CLAUDE.md, and the extension layer (skills, hooks, MCP, subagents, plugins, output styles, statusline).
---

# Configure Claude Code

This skill is the entry point for tuning Claude Code. For depth on a specific feature, jump to the dedicated skill (linked below). For the full reference, see the [official Claude Code docs](https://docs.claude.com/en/docs/claude-code/) — start with [Best Practices](https://code.claude.com/docs/en/best-practices), [How Claude Code Works](https://code.claude.com/docs/en/how-claude-code-works), and [Extend Claude Code](https://code.claude.com/docs/en/features-overview).

## Configuration Surface

Claude Code is configured through three layers:

| Layer | What it does | Where it lives |
|---|---|---|
| **`settings.json`** | Theme, model, permissions, hooks, statusline, env vars, plugins | `~/.claude/settings.json`, `.claude/settings.json`, `.claude/settings.local.json` |
| **`CLAUDE.md`** | Persistent instructions loaded every session | `~/.claude/CLAUDE.md`, `./CLAUDE.md`, `./CLAUDE.local.md` |
| **Extension layer** | Skills, subagents, hooks, MCP servers, output styles, plugins, slash commands | `~/.claude/{skills,agents,commands,output-styles}/`, `.claude/...`, `.mcp.json` |

### Settings file precedence

Lowest → highest priority. Higher levels override lower ones; `deny` rules at any level always win.

1. **User** — `~/.claude/settings.json` (this dotfiles repo, deployed via `mise run apply`)
2. **Shared project** — `.claude/settings.json` (committed)
3. **Local project** — `.claude/settings.local.json` (gitignored, secrets, per-clone overrides)
4. **Managed** — enterprise policies (cannot be overridden)

Permission arrays merge across scopes; most other keys override.

## Minimal `~/.claude/settings.json`

```jsonc
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "theme": "dark",
  "editorMode": "vim",
  "includeCoAuthoredBy": true,
  "autoUpdates": true,
  "cleanupPeriodDays": 7,
  "permissions": {
    "defaultMode": "auto"
  },
  "statusLine": {
    "type": "command",
    "command": "starship prompt",
    "padding": 0
  }
}
```

## Common Top-Level Keys

| Key | Type | Purpose |
|---|---|---|
| `theme` | `"dark"` / `"light"` / `"auto"` | TUI palette |
| `editorMode` | `"vim"` / `"emacs"` | Prompt editing keybindings |
| `model` | string | Default model (e.g., `"claude-opus-4-7"`); overridable with `/model` |
| `permissions` | object | `defaultMode`, `allow`, `ask`, `deny`, `additionalDirectories` |
| `hooks` | object | Lifecycle hooks — see `create-claude-hook` |
| `statusLine` | object | Custom status line — see `create-claude-statusline` |
| `outputStyle` | string | Active output style — see `create-claude-output-style` |
| `env` | object | Env vars exported into Claude Code's process |
| `enableAllProjectMcpServers` | bool | Whether to load `.mcp.json` servers without prompting |
| `enabledPlugins` | object | `{"plugin@marketplace": true}` — see `create-claude-plugin` |
| `includeCoAuthoredBy` | bool | Add `Co-Authored-By: Claude` to commits |
| `autoUpdates` | bool | Auto-update the CLI |
| `cleanupPeriodDays` | int | Auto-delete sessions older than N days from `~/.claude/projects/` |
| `spinnerTipsEnabled` | bool | Show rotating tips during waits |
| `preferredNotifChannel` | `"auto"` / `"terminal_bell"` / `"iterm2"` / `"notifications_disabled"` | Where to surface notifications |
| `skipAutoPermissionPrompt` | bool | Skip the first-run permission dialog |

Set values via `/config`, `/permissions`, or by editing the file directly. Run `claude config list` to inspect the resolved configuration.

### Permission modes

`permissions.defaultMode` accepts:

- `default` — ask before edits and shell commands
- `acceptEdits` — auto-accept file edits, still ask for shell
- `plan` — read-only tools only, produce a plan first
- `auto` — classifier model gates risky actions, routine work proceeds
- `bypassPermissions` — no prompts (dangerous; only inside isolated environments)

Cycle modes interactively with `Shift+Tab`.

### Permission rule syntax (allow / ask / deny)

```jsonc
{
  "permissions": {
    "allow": [
      "Bash(npm run *)",
      "Bash(git status:*)",
      "Read(./src/**)",
      "WebFetch(domain:github.com)",
      "mcp__github__*"
    ],
    "deny": [
      "Read(./.env)",
      "Read(~/.aws/**)",
      "Bash(curl *)"
    ]
  }
}
```

- `Bash(<pattern>)` — `*` wildcards, `:*` is a trailing-wildcard shorthand
- `Read(<path>)` / `Edit(<path>)` — gitignore-style globs; `~/...` for home, `//abs` for absolute
- `WebFetch(domain:<host>)` — apex + subdomains
- `mcp__<server>__<tool>` — specific MCP tools

For the full grammar (compound commands, wrapper handling, symlinks), see [Permissions](https://docs.claude.com/en/docs/claude-code/iam#permission-rule-syntax).

## CLAUDE.md (memory)

Loaded at the start of every session. Keep it under ~200 lines — bloated CLAUDE.md gets ignored.

| Location | Scope |
|---|---|
| `~/.claude/CLAUDE.md` | All sessions (this repo: `dot_claude/symlink_CLAUDE.md.tmpl`) |
| `./CLAUDE.md` | Project, committed |
| `./CLAUDE.local.md` | Project, gitignored |

Imports: `@path/to/file.md`. Run `/init` to scaffold a project CLAUDE.md.

What belongs there: bash commands Claude can't guess, code-style rules that differ from defaults, repo etiquette, env quirks, gotchas. What doesn't: anything Claude can read from the code itself.

## Extension Layer Quick Map

Each extension has a dedicated skill in this dotfiles repo:

| Need | Use | Skill |
|---|---|---|
| Reusable knowledge or workflow Claude can invoke | **Skill** | `create-agent-skill` |
| Slash command (single-shot prompt) | **Command** | `create-claude-command` |
| Specialized assistant with isolated context | **Subagent** | `create-claude-subagent` |
| Connect to external service (DB, API, browser) | **MCP server** | `create-claude-mcp-server` |
| Deterministic side effect on lifecycle event | **Hook** | `create-claude-hook` |
| Custom system prompt / persona | **Output style** | `create-claude-output-style` |
| Custom prompt status line | **Status line** | `create-claude-statusline` |
| Bundle and share the above | **Plugin** | `create-claude-plugin` |

Picking the right one ([cheat sheet](https://code.claude.com/docs/en/features-overview#match-features-to-your-goal)):

- **Always-on rule** → `CLAUDE.md`
- **On-demand knowledge or `/<name>` workflow** → skill
- **Heavy investigation that would bloat context** → subagent
- **Must run every time without asking** → hook
- **External service connection** → MCP
- **Reusable across many repos** → plugin

## Deployment in This Repo

Files in `dot_claude/` are managed by **chezmoi** and deployed to `~/.claude/` via `mise run apply`.

```
dot_claude/
├── settings.json              → ~/.claude/settings.json
├── symlink_CLAUDE.md.tmpl     → ~/.claude/CLAUDE.md (symlink)
├── agents/<name>.md           → ~/.claude/agents/<name>.md
├── commands/<name>.md         → ~/.claude/commands/<name>.md
└── skills/<name>/SKILL.md     → ~/.claude/skills/<name>/SKILL.md
```

After editing anything under `dot_claude/`, run **`mise run apply`** to deploy.

## Verification

- `/config` — interactive settings UI
- `/permissions` — review and edit permission rules
- `/hooks` — list configured hooks
- `/mcp` — list connected MCP servers and per-server token cost
- `/context` — see what's eating the context window
- `/doctor` — diagnose installation problems
- `claude config list` — print resolved config to stdout

## Guidelines

- **Keep `defaultMode` on `auto`** for low-friction work; rely on the classifier and explicit `deny` rules for guardrails.
- **Never put secrets in `dot_claude/settings.json`** — it's checked into a public dotfiles repo. Use `~/.claude/settings.local.json` (untracked) or env vars.
- **Audit `permissions.allow`** periodically — overly broad allowlists defeat the safety model.
- **Prune CLAUDE.md** whenever Claude starts ignoring rules — bloat is the usual cause.
- **Prefer hooks over prompt rules** for anything that must hold every time (e.g., "never edit `migrations/`").
