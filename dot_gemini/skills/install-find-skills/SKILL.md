---
name: install-find-skills
description: Install the find-skills meta-skill from vercel-labs/skills so the agent can discover and install other Agent Skills on demand.
---

# Install find-skills

[`vercel-labs/skills`](https://github.com/vercel-labs/skills) ships a **`find-skills`** meta-skill that teaches the agent how to discover, evaluate, and install other Agent Skills via `npx skills find` and the public skills leaderboard.

Once installed, the agent will automatically search for an existing skill before falling back to ad-hoc reasoning whenever a task matches a likely skill domain (e.g. "convert this React app to View Transitions" → suggest `vercel-labs/agent-skills/react-view-transitions`).

## When to Trigger

- The user wants the agent to suggest / pull skills proactively rather than memorizing the catalogue manually.
- The user mentions discovering Agent Skills, browsing the leaderboard, or "what skills exist for X".
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i find-skills`. If installed, skip.

## Install

```bash
# Install the find-skills meta-skill (project scope, default).
npx skills add vercel-labs/skills --skill find-skills

# Or interactively pick from the bundle.
npx skills add vercel-labs/skills
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## Companion: `npx skills find`

Even without the meta-skill installed, the same discovery flow is available interactively:

```bash
# Open an interactive picker for skills matching a query.
npx skills find "view transitions"

# JSON output for scripting.
npx skills find "react performance" --json
```

The meta-skill simply teaches the agent to invoke this command (and `npx skills add`) at the right moments.

## After Install

1. Restart the agent so progressive disclosure picks up the new skill description.
2. The agent will now suggest installing relevant skills before answering domain-heavy questions; approve or deny per request.
3. Project-scope installs (`.agents/skills/`) commit naturally. Global installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. The skill consults the public skills leaderboard — install counts and source reputation drive its picks.
2. It does **not** itself install skills automatically; it proposes, the agent approves, then `npx skills add` runs.
3. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [`vercel-labs/skills` repo](https://github.com/vercel-labs/skills)
- [find-skills SKILL.md](https://github.com/vercel-labs/skills/blob/main/skills/find-skills/SKILL.md)
- [Vercel changelog: skills v1.1.1 (interactive discovery)](https://vercel.com/changelog/skills-v1-1-1-interactive-discovery-open-source-release-and-agent-support)
- [Vercel KB: Agent Skills](https://vercel.com/kb/guide/agent-skills-creating-installing-and-sharing-reusable-agent-context)
