---
name: create-claude-subagent
description: Guide for creating new subagents for Claude Code
---

# Create Claude Code Subagent

This skill documents how to create a new Claude Code subagent. For more details, refer to the [official Claude Code subagents documentation](https://docs.claude.com/en/docs/claude-code/sub-agents).

## Subagent Directory Structure

Claude Code subagents are defined as Markdown files containing YAML frontmatter and system instructions. Subagents can be defined at two levels:

- **Local (Workspace):** `.claude/agents/` (project-specific subagents). **If the user does not specify where to create the subagent, assume it should be local.**
- **Global (chezmoi source of truth):** `~/.local/share/chezmoi/dot_claude/agents/` (personal subagents available across all workspaces after deployment).

Ask the user to run `mise run apply` to deploy global subagents to `~/.claude/agents/`.

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
description: <Short description of when to invoke this subagent>
tools: Read, Grep, Glob, Bash
model: sonnet
---

# <Agent Name Title>

You are the specialized <agent-name> agent. Your primary goal is to [describe primary capability]. Utilize your available tools precisely and autonomously to complete the user's request.
```

### Key Components

1. **Frontmatter:**
   - `name`: must match the filename (without `.md`); used to invoke the subagent.
   - `description`: a brief summary of *when* the orchestrator should delegate to this agent — phrase it as a trigger condition, not a feature list.
   - `tools` (optional): comma-separated list of tools the agent may invoke. Omit to inherit all tools available to the orchestrator. Restricting tools is the safer default.
   - `model` (optional): pin a specific Claude model (e.g., `sonnet`, `opus`, `haiku`). If omitted, the agent inherits the orchestrator's model.

1. **System Instruction:**
   - Everything after the second `---` is the prompt context provided to the subagent.
   - **Pattern:** Start with an H1 heading (e.g., `# Github Agent`).
   - **Persona:** Define its role clearly (e.g., `You are the specialized <agent-name> agent.`).
   - **Directive:** Give it a clear goal and explicit instructions on which tools to prefer and what success looks like.

## Step-by-Step Creation

1. **Create the file:** Create `.claude/agents/<name>.md` (or `~/.local/share/chezmoi/dot_claude/agents/<name>.md` if global).
2. **Fill the frontmatter:** Ensure the `name` matches the file name. Constrain `tools` to the minimum the agent needs.
3. **Draft the persona:** Keep the markdown instruction focused, clearly specifying the agent's responsibilities, allowed tools, and the expected output format.
4. **Test from the orchestrator:** Invoke the agent via the Task tool (or `/agents`) and verify it delegates correctly.
