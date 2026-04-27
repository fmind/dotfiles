---
name: install-vercel-skills
description: Install Vercel's official Agent Skills bundle for Web Interface Guidelines compliance, React composition patterns, React/Next.js performance, View Transitions, React Native, and Vercel deploys.
---

# Install Vercel Skills

Vercel publishes the official [`vercel-labs/agent-skills`](https://github.com/vercel-labs/agent-skills) bundle. The most universally useful entry is `web-design-guidelines` — a stack-agnostic UI/UX review pass. The rest of the bundle targets React / Next.js / Vercel deploys; install à la carte.

## When to Trigger

- The user asks for a UI / UX / accessibility review of any web interface (`web-design-guidelines`).
- The repo uses React / Next.js and the user wants performance, composition, or View Transitions guidance.
- The user is deploying to Vercel.
- The user is building a React Native / Expo app.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'web-design-guidelines|react-best-practices|view-transitions|composition-patterns|vercel-deploy'`. If installed, skip.

## Install

```bash
# List available skills.
npx skills add vercel-labs/agent-skills --list

# Whole bundle.
npx skills add vercel-labs/agent-skills

# Universal — UI review pass (recommended for any web project).
npx skills add vercel-labs/agent-skills --skill web-design-guidelines

# React / Next.js stack.
npx skills add vercel-labs/agent-skills \
  --skill react-best-practices \
  --skill composition-patterns \
  --skill react-view-transitions

# React Native / Expo.
npx skills add vercel-labs/agent-skills --skill react-native-guidelines

# Vercel deployment.
npx skills add vercel-labs/agent-skills --skill vercel-deploy-claimable
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

6 skills at the time of writing:

**Stack-agnostic**

- `web-design-guidelines` — Review UI code against Vercel's Web Interface Guidelines: layout, typography, motion, accessibility, perf hygiene. Useful for any web stack.

**React / Next.js**

- `react-best-practices` — Performance patterns from Vercel Engineering (data fetching, bundle splitting, RSC discipline).
- `composition-patterns` — Compound components, render props, context providers — React 19 API changes included.
- `react-view-transitions` — `<ViewTransition>`, `addTransitionType`, CSS view-transition pseudo-elements.

**Mobile**

- `react-native-guidelines` — React Native + Expo: list perf, animations, native modules.

**Deploy**

- `vercel-deploy-claimable` — Claimable Vercel deploy flow (interactive + token paths).

## Companion: `vercel-labs/skills` CLI

The CLI itself (`npx skills`) is published at [`vercel-labs/skills`](https://github.com/vercel-labs/skills). It also ships a meta-skill — `find-skills` — covered by the local `install-find-skills` skill.

## Dependencies

- Deploy skills assume the Vercel CLI: `npm i -g vercel` (not in mise by default).
- React skills assume a React/Next.js project layout.

## After Install

1. Restart the agent so progressive disclosure picks up the new descriptions.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. Update later via `npx skills update`.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate the content. Defer to the installed skills.
2. `web-design-guidelines` is the only universally useful entry for this dotfiles stack — the rest target Vercel/React/Next.js. Don't blanket-install if you're not on that stack.
3. There is overlap with `addyosmani/web-quality-skills` (covered by `install-web-quality-skills`): web-quality-skills focuses on Lighthouse pillars (Core Web Vitals, a11y, SEO), web-design-guidelines focuses on craft (layout, typography, motion). Both can coexist.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [`vercel-labs/agent-skills` repo](https://github.com/vercel-labs/agent-skills)
- [Vercel Web Interface Guidelines](https://vercel.com/design/guidelines)
- [`npx skills` CLI](https://github.com/vercel-labs/skills)
- [Vercel changelog: agent-skills launch](https://vercel.com/changelog/introducing-skills-the-open-agent-skills-ecosystem)
