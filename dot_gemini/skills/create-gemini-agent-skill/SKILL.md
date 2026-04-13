---
name: create-gemini-agent-skill
description: Guide for creating new skills for the Gemini CLI
---

# Create Gemini Agent Skill

This skill documents how to create functional skills (specialized files containing instructions) for the Gemini CLI. For more background on the conceptual model of skills, refer to the [official Gemini CLI Skills documentation](https://geminicli.com/docs/cli/skills/).

## Skill Directory Structure

Skills are defined as directories containing a main `SKILL.md` file and any related resources. Skills can be defined at two levels:

- **Local (Workspace):** `.gemini/skills/` (Project-specific skills). **If the
user does not specify where to create the skill, assume it should be local.**
- **Global (User):** `~/.gemini/skills/` (Personal skills available across all
workspaces).

_(Note: `.agent/skills/` and `~/.agent/skills/` are also supported as aliases)._

A skill directory typically looks like this:

```text
.gemini/skills/
  └── <skill-name>/
      ├── SKILL.md       # (Required) Metadata and instructions
      ├── scripts/       # (Optional) Executable scripts
      ├── references/    # (Optional) Static documentation
      └── assets/        # (Optional) Templates and other resources
```

## Skill File Structure (`SKILL.md`)

Each `SKILL.md` must start with YAML frontmatter specifying its identity, followed by detailed markdown instructions setting out the behavior, constraints, and operational steps for the agent to follow.

```markdown
---
name: <skill-name>
description: <Short description of the skill>
---

# <Skill Name Title>

This skill documents how to...

## Instructions

...
```

### Key Components

1. **Frontmatter:**
- `name`: A concise, hyphen-separated name for the skill (e.g., `create-gemini-agent-skill`). - `description`: A brief summary of what the skill helps the agent achieve.

2. **Body Content:**
- Define the exact steps the agent should follow when using this skill. - Document any project-specific conventions, code snippets, or directory structures. - Specify conditions, limitations, or constraints for executing the skill. - The richer the explanation and formatting, the better the Gemini agent can autonomously execute the procedures outlined within it.

## Step-by-Step Creation

1. **Create the folder:** Make the directory `.gemini/skills/<skill-name>` (or
`~/.gemini/skills/<skill-name>` if it must be global).
2. **Include optional folders:** If the skill requires them, create `scripts/`,
`references/`, or `assets/` subdirectories.
3. **Create the file:** Inside the skill folder, create `SKILL.md`.
4. **Fill the frontmatter:** Ensure you include the `name` and `description`
lines between the `---` delimiters.
5. **Draft the instructions:** Outline the procedural process explicitly,
following the markdown patterns of existing skills. Provide concrete examples and specific tool usage commands where relevant to anchor the logic firmly.
