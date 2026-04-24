---
name: reflect-on-conversation
description: Distil a Gemini CLI conversation into durable agents, skills, commands, or settings.
---

# Reflect on Conversation

Use this skill at the **end** of a productive session, when the user says "save this", "remember this", "make this permanent", or "reflect on what we did".

## Goal

Convert ephemeral, in-conversation knowledge into one or more of:

- A new **subagent** under `~/.gemini/agents/`.
- A new **skill** under `~/.gemini/skills/`.
- A new **slash command** under `~/.gemini/commands/`.
- A targeted edit to **`~/.gemini/settings.json`**.
- A note in the persistent memory store.

## Workflow

1. **Re-scan the conversation.** Identify:
   - Repeated intents (e.g. "every time I ask for a Python script you do X").
   - Friction points (extra confirmations, missing tools, wrong defaults).
   - One-off discoveries that have lasting value (a CLI flag, an API quirk).

1. **Classify each insight.** Use this decision tree:
   - Long-running specialised capability with its own MCP → **subagent**.
   - Multi-step workflow that isn't tied to a specific tool → **skill**.
   - Single repeatable prompt with deterministic output → **slash command**.
   - Behaviour change of the CLI itself → **`settings.json` patch**.
   - General fact worth remembering → **memory note**.

1. **Draft.** For each item, invoke the appropriate creation skill:
   - `create-gemini-subagent`
   - `create-gemini-agent-skill`
   - `create-gemini-command`
   - `setup-gemini-cli` for settings audits.

1. **Confirm.** Show the user the proposed diff, the file paths, and the exact
   prompt that would re-trigger the new artifact. Wait for approval.

1. **Persist.** Write the files. Append a one-line summary to
   `~/.gemini/REFLECTION.md` (create if missing) so the user can audit history.

## Guidelines

- Be ruthless about deduplication. If a similar agent/skill/command exists,
  *update* it instead of cloning.
- Never persist secrets. Reference env vars (`$GITHUB_MCP_PAT`, …) instead.
- Keep new artefacts small. A good agent is < 20 lines; a good skill < 80
  lines; a good command < 15 lines.
- If in doubt, prefer a **skill** (most flexible) over a subagent or command.
