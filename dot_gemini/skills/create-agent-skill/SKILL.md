---
name: create-agent-skill
description: Guide for creating new agent skills for Gemini CLI (cross-tool compatible)
---

# Create Agent Skill

Agent Skills are an open standard (originally by Anthropic, adopted by Gemini CLI, Claude Code, Cursor, OpenCode, and more — see [agentskills.io](https://agentskills.io)). Gemini CLI loads them via [progressive disclosure](https://geminicli.com/docs/cli/skills/): only the `name` and `description` are read at startup; the body is loaded only when a task matches.

## Discovery Locations

Gemini CLI searches three tiers (workspace > user > extension), and within each tier the `.agents/` alias takes precedence over `.gemini/`:

| Scope | Path | When to use |
|-------|------|-------------|
| Workspace | `.agents/skills/<skill-name>/` | Project-specific, version-controlled. |
| User | `~/.agents/skills/<skill-name>/` | Cross-workspace personal skills. |
| Chezmoi source | `~/.local/share/chezmoi/dot_gemini/skills/<skill-name>/` | Ask the user to run `mise run apply` to deploy to `~/.gemini/skills/`. |

**If the user does not specify a scope, default to creating the skill in `.agents/skills/<skill-name>/`** so it is committed alongside the project and works in other skills-aware tools.

## Skill Directory Layout

```text
<skill-name>/
├── SKILL.md       # (Required) frontmatter + instructions
├── scripts/       # (Optional) executable code
├── references/    # (Optional) static documentation the agent can grep
└── assets/        # (Optional) templates, prompts, fixtures
```

## `SKILL.md` Format

```markdown
---
name: <skill-name>
description: <One-line trigger that helps the agent decide when to load this skill>
---

# <Skill Title>

This skill documents how to...

## Instructions

...
```

Frontmatter — only two fields are required:

- `name` (required): slug — lowercase letters, digits, hyphens. Must match the folder name.
- `description` (required): Single sentence that lets the parent agent decide when this skill is relevant. The body is **not** loaded until the description matches.

The body is plain markdown: procedure, conventions, constraints, examples. Be concrete and tool-specific so the agent can act without re-prompting.

## Installing Skills From External Repos

```bash
# From a Git repo, into the current workspace (.gemini/skills/)
gemini skills install https://github.com/owner/repo.git --scope workspace

# Specific subdirectory of a monorepo
gemini skills install https://github.com/owner/repo.git --path skills/<skill-name>

# From a local path or a packaged .skill file
gemini skills install /path/to/skill
gemini skills install /path/to/my-skill.skill
```

The `.agents/skills/` alias is preferred for cross-tool compatibility.

## Step-by-Step Creation

1. **Pick a slug.** Lowercase, hyphenated, descriptive (e.g. `create-claude-hook`, `run-deep-research`).
2. **Create the folder.** `.agents/skills/<slug>/` for project-local (default), or `~/.local/share/chezmoi/dot_gemini/skills/<slug>/` for personal/global.
3. **Add optional subfolders** (`scripts/`, `references/`, `assets/`) only if the skill bundles resources.
4. **Write `SKILL.md`** with frontmatter (`name`, `description`) and a tight, procedural body. Keep the description action-oriented so progressive disclosure picks it up.
5. **Iterate** by trial-running the skill in the CLI; refine wording so Gemini reliably triggers and executes it.
