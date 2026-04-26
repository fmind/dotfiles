---
name: create-agent-skill
description: Guide for creating an Agent Skill that works in Claude Code, Gemini CLI, and other skills-aware tools.
---

# Create Agent Skill

Agent Skills are an open standard (originated by Anthropic, adopted by Claude Code, Gemini CLI, Cursor, OpenCode, and others — see [agentskills.io](https://agentskills.io)). A skill is a directory with a `SKILL.md` whose frontmatter is read at startup; the body is loaded only when the agent decides the skill is relevant (progressive disclosure).

In this dotfile setup, the user keeps a single shared library at `~/.agents/skills/` (chezmoi source: `dot_agents/skills/`). `~/.claude/skills/` and `~/.gemini/skills/` are symlinks to that directory, so any skill placed there is visible to both tools at once.

## Discovery Locations

When creating a skill, choose its scope:

| Scope | Path | When to use |
|-------|------|-------------|
| Workspace | `.agents/skills/<slug>/` | Project-specific, version-controlled with the repo. Both Gemini CLI and Claude Code (via the `.agents/` alias) discover it. |
| User (shared) | `~/.local/share/chezmoi/dot_agents/skills/<slug>/` | Personal skill available to **both** Claude Code and Gemini CLI after `mise run apply`. **This is the default for global skills.** |

If the user does not specify a scope, default to `.agents/skills/<slug>/` so the skill ships with the repo.

## Skill Directory Layout

```text
<slug>/
├── SKILL.md       # (required) frontmatter + instructions
├── scripts/       # (optional) executable code
├── references/    # (optional) static docs the agent can grep
└── assets/        # (optional) templates, prompts, fixtures
```

## `SKILL.md` Format

```markdown
---
name: <slug>
description: <one-line trigger that helps the agent decide when to load this skill>
---

# <Skill Title>

Short framing paragraph...

## Instructions

1. ...
2. ...
```

Frontmatter — only two fields are required:

- `name` (required): slug — lowercase letters, digits, hyphens. Must match the folder name.
- `description` (required): single sentence that lets the parent agent decide when this skill is relevant. The body is **not** loaded until the description matches, so make it concrete and trigger-rich.

The body is plain markdown: procedure, conventions, constraints, examples. Be tool-agnostic where possible. If the skill is genuinely tool-specific (e.g. uses a Claude-only or Gemini-only feature), name it accordingly (`create-claude-hook`, `use-google-workspace-cli`, …) so it is obvious from the slug.

## Step-by-Step Creation

1. **Pick a slug.** Lowercase, hyphenated, descriptive.
2. **Create the folder.** `.agents/skills/<slug>/` for project-local (default), or `~/.local/share/chezmoi/dot_agents/skills/<slug>/` for shared/global.
3. **Add optional subfolders** (`scripts/`, `references/`, `assets/`) only if the skill bundles resources.
4. **Write `SKILL.md`** with frontmatter (`name`, `description`) and a tight, procedural body.
5. **Deploy** (only for global): ask the user to run `mise run apply` so the new skill appears under `~/.claude/skills/` and `~/.gemini/skills/` (both are symlinks to `~/.agents/skills/`).
6. **Iterate** by trial-running the skill in either CLI; refine the description so progressive disclosure picks it up reliably.

## Cross-Tool Compatibility Tips

- Avoid hard-coded paths to one tool's config (e.g. `~/.claude/...`); prefer `~/.agents/...`.
- Avoid tool-specific frontmatter fields; stick to `name` and `description`.
- If the skill must call a tool-specific feature (e.g. Claude Code hooks, Gemini extensions), gate that in the body and document the requirement up front.
