---
name: create-cross-skill
description: Create a single Agent Skill in this chezmoi repo that is automatically visible to both Claude Code and Gemini CLI via the shared dot_agents/skills/ directory.
---

# Create Cross-Tool Agent Skill

Use this skill when adding a new Agent Skill to this dotfiles repo that should work in **both** Claude Code and Gemini CLI. Unlike commands, Agent Skills share an open standard (`SKILL.md` with frontmatter — see [agentskills.io](https://agentskills.io)), so we only need a **single** source file. The dotfile layout already symlinks both tools' skill directories at `~/.agents/skills/`.

## Layout in this repo

```
dot_agents/skills/<slug>/SKILL.md     ← canonical, single source of truth
dot_claude/symlink_skills.tmpl        → ~/.agents/skills (created on chezmoi apply)
dot_gemini/symlink_skills.tmpl        → ~/.agents/skills (created on chezmoi apply)
```

After `mise run apply`:

- `~/.claude/skills/` is a symlink to `~/.agents/skills/`
- `~/.gemini/skills/` is a symlink to `~/.agents/skills/`

So a single skill in `dot_agents/skills/<slug>/` is automatically discovered by both tools.

## When NOT to use this skill

If the skill genuinely depends on a tool-specific feature, do **not** make it cross-tool — name it explicitly so the slug makes the dependency obvious:

- `create-claude-hook`, `create-claude-mcp-server` → Claude Code only
- `use-google-workspace-cli`, `use-google-agents-cli` → Gemini-leaning workflows

These still live under `dot_agents/skills/` (single library) but their slug warns the orchestrator off when not relevant.

## `SKILL.md` Format

```markdown
---
name: <slug>
description: <one-sentence trigger that lets the agent decide when to load this skill>
---

# <Skill Title>

Short framing paragraph — what the skill does and when to invoke it.

## Instructions

1. ...
2. ...
3. ...

## Guidelines

- ...
```

Frontmatter — only two fields are required and both tools recognise them:

- `name` (required): slug — lowercase letters, digits, hyphens. Must match the folder name.
- `description` (required): single sentence used for **progressive disclosure** — the body is loaded only when this matches the task. Make it concrete and trigger-rich.

## Step-by-Step

1. **Pick a slug.** Lowercase, hyphenated, descriptive (`create-cross-command`, `run-deep-research`, ...). Avoid `claude-` or `gemini-` prefixes unless the skill is genuinely tool-specific.
2. **Create the folder** `dot_agents/skills/<slug>/`.
3. **Add optional subfolders** only if the skill bundles resources:
   - `scripts/` — executable code the agent can run
   - `references/` — static docs the agent can grep
   - `assets/` — templates, prompts, fixtures
4. **Write `SKILL.md`** with frontmatter (`name`, `description`) and a tight, procedural body. Stay tool-agnostic in the wording — refer to "the agent" or "the CLI", not "Claude" or "Gemini", unless a tool-specific feature is unavoidable.
5. **Deploy**: ask the user to run `mise run apply`. The skill is then visible in both `~/.claude/skills/` and `~/.gemini/skills/` via the shared symlink.

## Cross-Tool Compatibility Tips

- **Avoid hard-coded tool paths** in the body (`~/.claude/...`, `~/.gemini/...`). Prefer `~/.agents/...` or relative paths.
- **Stick to `name` and `description`** in frontmatter. Tool-specific extra fields are silently ignored by the other tool, but they add noise.
- **Be concrete** — both tools use the description for progressive disclosure. Vague descriptions (`Helps with code`) won't trigger reliably.
- **If the skill bundles scripts**, make sure they don't assume one tool's environment (e.g. don't depend on `$CLAUDE_PROJECT_DIR`).

## Reference: existing cross-tool skills in this repo

`create-agent-skill`, `create-python-script` — both live in `dot_agents/skills/` and work in both tools. Use them as templates.
