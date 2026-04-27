---
name: setup-gemini-cli-on-new-project
description: Bootstrap Gemini CLI on a fresh repo — create the `.gemini/` folder layout, write `settings.json`, generate `GEMINI.md` via `/init`, decide what to commit vs `.gitignore`, and link to follow-up skills (hooks, memory, MCP servers, subagents, commands).
---

# Setup Gemini CLI on a New Project

This skill walks through the first 10 minutes of using Gemini CLI on a repo that has never had a `.gemini/` folder. The goal is a committed `.gemini/` directory that any teammate (or fresh checkout on another machine) can use immediately.

There is **no `gemini init` shell command**; the workflow uses the in-session `/init` slash command plus a hand-authored `settings.json`.

## Prerequisites

- `gemini` is installed and you can authenticate (`/auth` if you need to switch later — see `use-gemini-cli`).
- The repo has a clean working tree, or at least no untracked `.gemini/` from a stale experiment.
- You're inside the project root (`cd <repo>`).

## Quickstart (5 minutes)

```bash
mkdir -p .gemini/{commands,agents,skills,hooks}
touch .geminiignore
$EDITOR .gemini/settings.json     # paste the starter below
gemini                            # accept the trust prompt for this folder
# inside the session:
/init                             # generates GEMINI.md tailored to this repo
/memory list                      # confirm context files are loaded
/permissions                      # confirm trust + shell allow/deny
exit
git add .gemini AGENTS.md GEMINI.md .geminiignore
```

## Recommended `.gemini/` Layout

```text
.gemini/
├── settings.json          # workspace overrides (commits with repo)
├── GEMINI.md              # Gemini-specific persona (optional; root-level is fine too)
├── commands/              # custom slash commands (.toml)
│   └── git/commit.toml
├── agents/                # subagents (.md, frontmatter uses mcp_servers: snake_case)
│   └── reviewer.md
├── skills/                # workspace-scope agent skills (auto-memory promotion target)
│   └── house-style/SKILL.md
├── hooks/                 # hook scripts referenced from settings.json
│   ├── block-secrets.sh
│   └── format.sh
└── plans/                 # plan-mode artifacts (path set in settings.json)
```

Plus repo-root files:

```text
AGENTS.md                  # tool-agnostic project rules (Gemini, Claude, Cursor, ...)
GEMINI.md                  # optional: Gemini-only overrides if not in .gemini/
.geminiignore              # paths to hide from the agent (independent of .gitignore)
```

## Starter `.gemini/settings.json`

A minimal but useful baseline. Strip what you don't need.

```json
{
  "context": {
    "fileName": ["AGENTS.md", "GEMINI.md"]
  },
  "general": {
    "checkpointing": { "enabled": true },
    "defaultApprovalMode": "auto_edit",
    "plan": { "directory": ".gemini/plans" }
  },
  "experimental": {
    "autoMemory": true,
    "jitContext": true
  },
  "tools": {
    "shell": { "enableInteractiveShell": true }
  },
  "mcpServers": {
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          { "name": "ctx", "type": "command", "command": "$GEMINI_PROJECT_DIR/.gemini/hooks/inject-context.sh", "timeout": 5000 }
        ]
      }
    ]
  }
}
```

Notes:

- **No comments allowed** in `settings.json` (strict JSON). Document choices in a sibling `.gemini/README.md` if you need to.
- `mcpServers` is **camelCase** in `settings.json`. Subagent frontmatter uses `mcp_servers:` (snake_case) — see `create-gemini-subagent`.
- `defaultApprovalMode: "auto_edit"` lets the agent write files without prompting but still asks for shell commands.

## Starter `AGENTS.md`

```markdown
# Project Agent Rules

`<project-name>` — `<one-line description>`.

## House rules
- Stack: <languages, frameworks>.
- Test runner: `<cmd>` — run before declaring a task done.
- Lint: `<cmd>`. Format: `<cmd>`.
- Secrets: never commit; use `.env.local` (gitignored) and `<secret manager>`.
- Branching: `<convention>`. Commit style: `<convention>`.

## Source layout
- `src/` — app code.
- `tests/` — unit + integration tests (mirror `src/` paths).
- `infra/` — IaC (terraform / pulumi / ...).
- `.gemini/` — Gemini CLI workspace config (commits).

## Editing workflow
1. Branch from `main`: `git switch -c <feat>/...`.
2. Run `<test cmd>` and `<lint cmd>` before opening the PR.
3. Don't commit on the agent's behalf — surface the diff and let the human commit.
```

## Starter `.geminiignore`

```gitignore
# Generated / vendored — agent doesn't need to read these.
**/generated/**
**/__pycache__/**
**/node_modules/**
**/*.lock
**/dist/**
**/build/**

# Large or sensitive fixtures.
tests/fixtures/large/**
```

## What to Commit vs Gitignore

**Commit (so teammates inherit the setup):**

- `.gemini/settings.json`
- `.gemini/commands/`, `.gemini/agents/`, `.gemini/skills/`, `.gemini/hooks/`
- `AGENTS.md`, `GEMINI.md`, `.geminiignore`

**Add to `.gitignore`:**

```gitignore
# Local-only Gemini CLI state
.gemini/plans/
.gemini/tmp/
.gemini/.local*
.gemini/settings.local.json
```

User-global state lives in `~/.gemini/` and never enters the repo (sessions, transcripts at `~/.gemini/tmp/<project>/chats/`, personal `GEMINI.md`).

## First-launch Trust

The first `gemini` invocation in a new repo prompts: **Trust folder / Trust parent / Deny**. The choice is persisted to `~/.gemini/trustedFolders.json`.

- **Untrusted = safe mode**: project `settings.json`, MCP servers, hooks, custom commands, extensions, and auto-memory are all disabled.
- **Trusted**: full project config kicks in.
- For one-off CI runs, pass `gemini --skip-trust` instead of trusting permanently.

## Generate `GEMINI.md` with `/init`

Inside the first session, run:

```text
/init
```

The CLI inspects the cwd (languages, build tools, README, package.json, etc.) and writes a tailored `GEMINI.md` you can then trim. Treat it as a starting draft, not a finished file — review every section.

## Verify the Setup

Inside the session:

```text
/memory list                     show every context file currently loaded
/memory show                     dump the concatenated buffer (debug)
/permissions                     workspace trust + shell allow/deny + tool gates
/mcp list                        any project MCP servers connected
/agents list                     project subagents discovered
/extensions                      active extensions for this run
```

From the shell:

```bash
gemini extensions list
gemini mcp list
gemini skills list
python -m json.tool .gemini/settings.json    # validate JSON
```

## Sharing With Teammates

Two paths — pick per workflow:

1. **Commit `.gemini/`** (this skill's default). Simplest; setup applies the moment a teammate clones and accepts the trust prompt.
2. **Ship a Gemini CLI extension**. Wraps MCP servers + commands + agents + skills + hooks into a redistributable bundle that lives outside the repo. Use this when the same setup should travel across many repos. See `configure-gemini-extensions`.

In most cases, commit `.gemini/`. Reach for an extension only when the same scaffold needs to apply to ≥3 repos.

## Follow-up Skills

After the initial scaffold, deepen each surface:

| Surface | Skill |
|---------|-------|
| Settings schema, MCP servers, trust | `configure-gemini-cli` |
| Hooks (block secrets, auto-format, session-start context) | `configure-gemini-cli-hooks` |
| `GEMINI.md` / `AGENTS.md`, `.geminiignore`, auto-memory | `configure-gemini-cli-memory` |
| Custom slash commands | `create-gemini-command` |
| Subagents (remember: `mcp_servers:` snake_case in frontmatter) | `create-gemini-subagent` |
| Workspace-scope skills | `create-agent-skill` |
| Vendor MCP packs (Firebase, BigQuery, GitHub, ...) | `install-*-mcp` |
| Vendor skill packs (Firebase, Stitch, Angular, ...) | `install-*-skills` |
| CLI flags, headless mode, slash commands, sessions | `use-gemini-cli` |
| Extensions vs project config | `configure-gemini-extensions` |

## Important Notes

1. **Run `/init` only after committing the rest of the scaffold.** It writes `GEMINI.md` from a model draft — easier to review as an isolated diff.
2. **`mcpServers` (settings.json, camelCase) and `mcp_servers` (subagent frontmatter, snake_case) are different on purpose.** Mixing them up silently disables the MCP layer in subagents.
3. **`.gemini/plans/` and `.gemini/tmp/` belong in `.gitignore`** — they're per-session scratch.
4. **The trust prompt is per-folder, per-machine.** Every fresh checkout (or worktree) needs to re-trust once.
5. **`settings.json` is strict JSON** — no comments, no trailing commas. Use `python -m json.tool` to validate.
6. **Don't put secrets in `settings.json`.** Use `"$VAR"` substitution or a secret manager; `settings.json` is committed.
7. **Workspace settings override user settings.** A repo can pin a model, approval mode, or MCP allowlist that supersedes the global default for that folder.

## Documentation

- [Settings & precedence](https://geminicli.com/docs/cli/settings/)
- [Trusted folders](https://geminicli.com/docs/cli/trusted-folders/)
- [Slash commands reference (`/init`, `/memory`, ...)](https://geminicli.com/docs/reference/commands/)
- [Configuration reference](https://geminicli.com/docs/reference/configuration/)
- [Extensions overview](https://geminicli.com/docs/extensions/)
- Companion skills: `configure-gemini-cli`, `configure-gemini-cli-hooks`, `configure-gemini-cli-memory`, `configure-gemini-extensions`, `create-gemini-command`, `create-gemini-subagent`, `create-agent-skill`, `use-gemini-cli`.
