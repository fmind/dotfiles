---
name: install-google-stitch-skills
description: Install Google Stitch skills bundle for AI-driven UI design (screens, React components, design systems, walkthrough videos). Use for Stitch UI projects.
---

# Install Google Stitch Skills

Google Labs publishes the official [`google-labs-code/stitch-skills`](https://github.com/google-labs-code/stitch-skills) bundle. It's the canonical source for working with [Google Stitch](https://stitch.withgoogle.com) (powered by Gemini 3 Pro / Flash) — turning natural-language prompts, sketches, and screenshots into polished UI designs and frontend code.

## When to Trigger

- The user wants to design app screens, multi-screen flows, components, or design systems from prompts / sketches / screenshots.
- The user mentions Stitch, "design-to-code", `DESIGN.md` generation, prompt enhancement for UI, or shadcn/ui from Stitch.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'stitch|design-md|enhance-prompt|taste-design'`. If installed, skip.

## Install

```bash
# List available skills.
npx skills add google-labs-code/stitch-skills --list

# Install one (repeat per skill — project scope by default).
npx skills add google-labs-code/stitch-skills --skill stitch-design --project
npx skills add google-labs-code/stitch-skills --skill react-components --project
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted. Cross-tool: Claude Code, Gemini CLI, Cursor, and Antigravity all read `.agents/skills/`.

## What Gets Installed

8 skills at the time of writing:

- **stitch-design** — Unified prompt-enhancement + screen generation entry point.
- **stitch-loop** — Multi-page website from a single prompt.
- **design-md** — Generates `DESIGN.md` documenting the design system.
- **enhance-prompt** — Stitch-optimized prompt rewriting.
- **react-components** — Stitch screens to React component systems with token validation.
- **remotion** — Walkthrough videos of Stitch projects via Remotion.
- **shadcn-ui** — shadcn/ui integration guidance.
- **taste-design** — Semantic anti-slop design system generator.

## Related: Stitch MCP server

The skills drive the Stitch MCP server at `https://stitch.googleapis.com/mcp`. Two supported auth methods (pick one):

- **API key** (recommended for individuals, non-expiring) — generate in Stitch settings, send as the `X-Goog-Api-Key` header.
- **OAuth / Application Default Credentials** (1-hour tokens, stdio proxy refreshes):

  ```bash
  gcloud auth application-default login
  gcloud services enable stitch.googleapis.com --project=<PROJECT_ID>
  ```

Up to 5 interconnected screens can be generated as a flow. MCP tools include: list/get projects, list/get screens, download images and HTML, generate new screens, enhance prompts.

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. The `stitch-design` skill is the natural entry point — start there, then drill into `react-components` / `shadcn-ui` for code generation, or `design-md` / `taste-design` for design-system work.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate Stitch docs. Defer to the installed skills.
2. Stitch generation consumes quota on Google's side — be deliberate about regenerating screens during iteration.
3. Export targets are Figma (preserves layers/components), React, or HTML/CSS. The `react-components` skill enforces design-token validation on export.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [Google Stitch](https://stitch.withgoogle.com)
- [Stitch docs](https://stitch.withgoogle.com/docs/)
- [`google-labs-code/stitch-skills` repo](https://github.com/google-labs-code/stitch-skills)
- [MCP setup](https://stitch.withgoogle.com/docs/mcp/setup/)
- [MCP guide](https://stitch.withgoogle.com/docs/mcp/guide/)
- [Gemini CLI extension](https://github.com/gemini-cli-extensions/stitch)
- [Codelab — Design-to-Code with Antigravity + Stitch](https://codelabs.developers.google.com/design-to-code-with-antigravity-stitch)
