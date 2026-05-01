---
name: create-gemini-cli-subagent
description: Guide for creating Gemini CLI subagents — frontmatter (name, description, kind, tools, `mcp_servers` snake_case), persona pattern, scope (workspace vs global via chezmoi), future-proofing rules, and silent-failure pitfalls.
---

# Create Gemini CLI Subagent Skill

This skill documents how to create a new Gemini CLI subagent. For more details, refer to the [official Gemini CLI subagent documentation](https://geminicli.com/docs/core/subagents/).

## Subagent Directory Structure

Gemini subagents are defined as Markdown files containing YAML frontmatter and system instructions. Subagents can be defined at two levels:

- **Local (Workspace):** `.gemini/agents/` (Project-specific subagents). **If the user does not specify where to create the subagent, assume it should be local.**
- **Global (chezmoi source of truth):** `~/.local/share/chezmoi/dot_gemini/agents/` (Personal subagents available across all workspaces after deployment).

Ask the user to run `mise run apply` to deploy global subagents to `~/.gemini/agents/`.

Example location:

```text
<agents-root>/
  └── <agent-name>.md
```

## Subagent File Structure

Each subagent file follows a standard format:

```markdown
---
name: <agent-name>
description: <Short description of the agent>
kind: local
tools:
  - "*"
mcp_servers:
  <server-name>:
    # MCP Server configuration goes here (httpUrl, command/args, env, headers, etc.)
---

# <Agent Name Title>

You are the specialized <agent-name> agent. Your primary goal is to [describe primary capability]. Utilize your available tools precisely and autonomously to complete the user's request.
```

### Key Components

1. **Frontmatter (must use these exact field names):**
   - `name` (required): Must match the filename (without `.md`); slug only — lowercase letters, digits, hyphens, underscores.
   - `description` (required): Short summary used by the parent agent to decide when to delegate.
   - `kind` (optional): `local` (default) or `remote`.
   - `tools` (optional): Allowlist of tool names. Use `"*"` to inherit every parent tool; otherwise list each tool explicitly. Omitted means inherit-all.
   - `mcp_servers` (optional, **snake_case** — `mcpServers` is silently ignored in subagent frontmatter, even though the top-level `settings.json` uses camelCase): Inline MCP server definitions isolated to this agent. Each entry takes the standard MCP transport keys (`httpUrl`, `command`/`args`, `env`, `headers`, `authProviderType`, …).
   - `model`, `temperature`, `max_turns`, `timeout_mins` (optional): Tune the underlying LLM call.

2. **System Instruction:**
   - Everything after the second `---` is the prompt context provided to the subagent.
   - **Pattern:** Start with an H1 heading (e.g., `# Github Agent`).
   - **Persona:** Define its role clearly (e.g., `You are the specialized <agent-name> agent.`).
   - **Directive:** Give it a clear goal and the instruction to `Utilize your available tools precisely and autonomously...`.

## Step-by-Step Creation

1. **Create the file:** Create `.gemini/agents/<name>.md` (or `~/.local/share/chezmoi/dot_gemini/agents/<name>.md` if global).
2. **Fill the frontmatter:** Ensure `name` matches the filename. Use `mcp_servers` (snake_case!) for any inline MCP server config — the camelCase spelling is silently ignored in subagent frontmatter.
3. **Draft the persona:** Keep the markdown instruction focused, clearly specifying the agent's responsibilities inline with the established standard format.

## Future-Proofing Rules

Subagent files outlive the moment they were written — Gemini CLI versions, model generations, and product names all turn over faster than agent prose gets re-edited. Write each agent so it stays correct without maintenance.

- [ ] **No time-stamped prose.** Avoid "As of 2026…", "announced at Cloud Next 2026", "Q4 2025", "currently", "recently", "new in v…". Write capabilities in plain present tense and let linked docs carry the timestamp.
- [ ] **No rebrand narration.** Don't write "rebranded from X", "formerly known as Y", "the new name for Z". Use the current product name; if disambiguation is genuinely useful (e.g. "Gemini Cloud Assist (NOT Gemini Code Assist)"), keep it to a one-liner aimed at *clarification*, not history.
- [ ] **No specific model versions in prose.** Don't pin "powered by Gemini 3 Pro / Flash" or "uses Claude 4.5 Sonnet". Reference model families abstractly ("Gemini Pro/Flash family") or omit — pinning a generation locks the file to that generation.
- [ ] **Don't pin frontmatter tuning fields (`model`, `temperature`, `max_turns`, `timeout_mins`)** unless tuning is the agent's explicit purpose. These are harness/CLI-controlled defaults; pinning them in an agent file ages with each Gemini CLI release.
- [ ] **Pin executable references; don't assume global installs.** In `mcp_servers`, prefer `npx -y <pkg>@latest` or `uvx --from <pkg>` over `command: <bare-binary>`. A globally-installed CLI on the user's PATH drifts independently of the agent file. For Docker, pin the image tag explicitly.
- [ ] **Cite canonical product docs, not announcements.** Link to `cloud.google.com/<product>/docs` and similar evergreen paths. Avoid blog posts, release notes, codelabs, and "introducing X" URLs whose paths rot.
- [ ] **Describe the agent's *purpose*, not the product's *current state*.** Capabilities and workflows describe what the agent does; they don't need to narrate where the product sits in its lifecycle. If a feature is in preview, say "preview" — don't say "newly released in preview".

## Common Pitfalls

Check each item before shipping a new subagent — these are silent failures (the agent loads but misbehaves):

- [ ] **`mcp_servers` is snake_case.** The camelCase `mcpServers` form (used in top-level `settings.json`) is silently ignored in subagent frontmatter — the agent loads with no tools and no error.
- [ ] **`name:` matches the filename slug** (without `.md`). Mismatch breaks delegation routing.
- [ ] **`description:` is a routing trigger, not a label.** Write "Use for X, Y, Z" so the parent agent knows *when* to delegate. Avoid static descriptions like "X agent for Y".
- [ ] **Persona uses the human title, not the slug.** Write "specialized Cloud Storage agent", not "specialized cloud-storage agent".
- [ ] **Only one MCP server unless intentional.** Registering two servers for the same purpose (e.g. `_http` + `_local` fallback) doubles the tool count and confuses tool selection.
- [ ] **Inline MCP vs `~/.gemini/settings.json`.** Inline `mcp_servers` is isolated to the subagent. Put servers in `settings.json` only when multiple agents (or the root agent) need them — otherwise prefer inline for scope-locality.
- [ ] **Don't include host-specific instructions** (e.g. "Claude Code MCP setup") inside a Gemini CLI subagent file.

## Documentation

- [Gemini CLI subagents](https://geminicli.com/docs/core/subagents/)
- [MCP server configuration](https://geminicli.com/docs/tools/mcp-server/)
