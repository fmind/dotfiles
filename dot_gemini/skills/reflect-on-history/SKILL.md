---
name: reflect-on-history
description: Reflect on recent Gemini CLI conversations to distill durable agents, skills, commands, or settings.
---

# Reflect on History

Use this skill when the user wants to reflect on recent sessions or reflect on what we've been doing lately.

## Goal

Aggregate insights from multiple recent conversations into one or more of:

- A new **subagent** under `~/.local/share/chezmoi/dot_gemini/agents/`.
- A new **skill** under `~/.local/share/chezmoi/dot_gemini/skills/`.
- A new **slash command** under `~/.local/share/chezmoi/dot_gemini/commands/`.
- A targeted edit to `~/.local/share/chezmoi/dot_gemini/settings.json`.

Ask the user to run `chezmoi apply` to deploy those global artifacts to the default Gemini path under `~/.gemini/`.

## Workflow

1.  **Locate recent session files.**
    - Scan `~/.gemini/tmp/` recursively for chat session files.
    - Use `fd` (modern find) to filter for files modified within the last 24 hours (default):
      ```bash
      fd --extension json --extension jsonl --changed-within 24h --search-path ~/.gemini/tmp/
      ```
    - If `fd` is not available, fall back to `find`:
      ```bash
      find ~/.gemini/tmp/ -name "*.json*" -mtime -1
      ```

2.  **Read and aggregate.**
    - Read the contents of the identified session files.
    - Identify:
        - Repeated intents across different sessions.
        - Persistent friction points or missing capabilities.
        - One-off discoveries that have been useful in multiple contexts.

3.  **Classify each insight.** Use this decision tree:
    - Long-running specialised capability with its own MCP → **subagent**.
    - Multi-step workflow that isn't tied to a specific tool → **skill**.
    - Single repeatable prompt with deterministic output → **slash command**.
    - Behaviour change of the CLI itself → **`settings.json` patch**.
    - General fact worth remembering → **memory note**.

4.  **Draft.** For each item, invoke the appropriate creation skill:
    - `create-gemini-command`
    - `create-gemini-subagent`
    - `create-gemini-agent-skill`

5.  **Confirm.** Show the user the proposed diff, the file paths, and the exact prompt that would re-trigger the new artifact. Wait for approval.

## Guidelines

- **Context Efficiency:** If there are many recent sessions, summarize the key takeaways rather than reading every single line of every file to avoid hitting context limits.
- **Deduplication:** If a similar agent/skill/command exists, *update* it instead of cloning.
- **Never persist secrets:** Reference env vars (`$GITHUB_MCP_PAT`, etc.) instead.
- **Keep it small:** New artefacts should be concise and focused.
- **If in doubt, prefer a skill** (most flexible) over a subagent or command.
