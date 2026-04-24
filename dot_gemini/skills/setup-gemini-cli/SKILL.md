---
name: setup-gemini-cli
description: End-to-end skill to configure Gemini CLI — settings, agents, skills, commands, MCPs, and API keys.
---

# Setup Gemini CLI

This skill is the **master configuration skill**. Use it when the user asks to "configure Gemini CLI", "set me up", "wire all the agents", or to reflect on a previous conversation and turn the result into persistent configuration.

## Layout

Configuration lives under `~/.gemini/` (global) or `.gemini/` (workspace):

```text
~/.gemini/
├── settings.json        # global settings (theme, vim mode, MCP defaults)
├── agents/<name>.md     # subagents (one per integration)
├── skills/<name>/       # SKILL.md + optional scripts/, references/, assets/
└── commands/<name>.toml # slash-command shortcuts
```

## Workflow

1. **Inventory.** Read every file under `~/.gemini/` first (use `read_file`).
   Identify what already exists; never overwrite without diffing first.

1. **Audit `settings.json`.** Confirm the following keys are present and
   reasonable. If anything is missing, propose a unified diff before editing.

   - `general.preferredEditor`: matches the user's editor (this repo: `nvim`).
   - `general.vimMode`: `true` (matches user's global vim preference).
   - `general.checkpointing.enabled`: `true` for safe rollbacks.
   - `general.defaultApprovalMode`: `auto_edit` for productive flow.
   - `ui.theme`: `catppuccin-mocha` (consistent across this dotfiles repo).
   - `ui.hideBanner` / `ui.hideTips`: `true` for a quiet TUI.
   - `security.auth.selectedType`: `oauth-personal` unless the user uses an
     API key.
   - `experimental.*`: enable `contextManagement`, `jitContext`,
     `memoryManager`, `taskTracker`, `worktrees`.

1. **Audit agents.** For each integration the user mentions, ensure a
   subagent exists under `agents/`. Use the `create-gemini-subagent` skill to
   draft new ones. Cross-check MCP endpoints against
   <https://docs.cloud.google.com/mcp/supported-products>.

1. **Audit skills.** For each repeatable workflow the user describes, ensure
   a skill exists under `skills/`. Use the `create-gemini-agent-skill` skill to
   draft new ones.

1. **Audit commands.** For one-shot prompts the user runs frequently
   (`/commit`, `/review`, `/explain`), ensure a `commands/<name>.toml` exists.
   Use the `create-gemini-command` skill.

1. **API keys.** Verify the user knows which environment variables their
   chosen agents require (e.g. `GITHUB_MCP_PAT`, `STITCH_ACCESS_TOKEN`,
   `JULES_API_KEY`, `GEMINI_API_KEY`). Suggest storing them in a fish-private
   file sourced by `~/.private.fish` rather than committing them.

1. **Validate.** Run `gemini --help` and `gemini list-agents` (if available).
   Surface any errors and remediate.

## Reflection mode

When the user asks you to *reflect on the previous conversation* and persist
the lessons:

1. Re-read the conversation transcript or recent shell history.
1. Identify recurring intents that are not yet covered by an agent / skill /
   command. Propose new ones explicitly.
1. Identify settings that caused friction (e.g. excessive confirmations).
   Propose targeted edits to `settings.json`.
1. Present a summary diff and wait for approval before writing files.

## Guidelines

- Never write secrets to disk inside `~/.gemini/`. Use environment variables.
- Keep agent files short — a strong persona + the MCP server config is enough.
- Every new skill must be testable with a 1-sentence prompt.
- Default to global (`~/.gemini/`) unless the user asks for a workspace scope.
