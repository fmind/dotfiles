---
name: create-gemini-subagent
description: Guide for creating new subagents for the Gemini CLI
---

# Create Gemini Subagent Skill

This skill documents how to create a new Gemini CLI subagent. For more details, refer to the [official Gemini CLI subagent documentation](https://geminicli.com/docs/core/subagents/).

## Subagent Directory Structure

Gemini subagents are defined as Markdown files containing YAML frontmatter and system instructions. Subagents can be defined at two levels:

- **Local (Workspace):** `.gemini/agents/` (Project-specific subagents). **If
the user does not specify where to create the subagent, assume it should be local.**
- **Global (User):** `~/.gemini/agents/` (Personal subagents available across
all workspaces).

Example location:

```text
.gemini/agents/
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
  <agent-name>:
    # MCP Server configuration goes here
---

# <Agent Name Title>

You are the specialized <agent-name> agent. Your primary goal is to [describe
primary capability]. Utilize your available tools precisely and autonomously to
complete the user's request.
```

### Key Components

1. **Frontmatter:**
- `name`: Must match the filename (without `.md`). - `description`: A brief summary of what the agent does. - `kind`: `local` (default) or remote. - `tools`: A list of tools the agent can access. Can be explicit tool names, or wildcards like `*` (all tools), `mcp_*` (all MCP tools), or `mcp_<server-name>_*` (tools from a specific MCP server). - `mcp_servers`: Configuration for inline Model Context Protocol servers unique to the agent.

2. **System Instruction:**
- Everything after the second `---` is the prompt context provided to the subagent. - **Pattern:** Start with an H1 heading (e.g., `# Github Agent`). - **Persona:** Define its role clearly (e.g., `You are the specialized <agent-name> agent.`). - **Directive:** Give it a clear goal and the instruction to `Utilize your available tools precisely and autonomously...`.

## Step-by-Step Creation

1. **Create the file:** Create `.gemini/agents/<name>.md` (or
`~/.gemini/agents/<name>.md` if global).
2. **Fill the frontmatter:** Ensure the `name` matches the file name, and use
one of the `mcp_servers` patterns above.
3. **Draft the persona:** Keep the markdown instruction focused, clearly
specifying the agent's responsibilities inline with the established standard format.
