---
name: astro
description: "Build content-driven websites and web apps with Astro: islands, Tailwind CSS v4, Content Layer, Vitest, Prettier, and mise tasks. Use for Astro sites, SSG, SSR, or islands."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/astro
  created: "2026-09-04"
  updated: "2026-09-04"
---

# Astro Standard

Astro builds content-driven websites and web apps with islands architecture and zero JavaScript by default. Shared TypeScript conventions come from [typescript-stack](../typescript-stack/SKILL.md); documentation sites can also evaluate [hugo](../hugo/SKILL.md); deploy targets from [firebase](../firebase/SKILL.md) and [cloud-run](../cloud-run/SKILL.md).

## 1. Core Stack

- **CLI & framework**: `astro` via project scripts (`pnpm astro`, `pnpm run build`) or `pnpm dlx create-astro@latest`; every project pins `astro` in `package.json` with `pnpm`.
- **Architecture**: zero JavaScript by default, partial hydration islands (`client:load`, `client:idle`, `client:visible`, `client:only`), Server Islands (`server:defer`), Content Layer API.
- **Styling**: Tailwind CSS v4 via `@tailwindcss/vite` plugin in [astro.config.mjs](references/astro.config.mjs) and `@import "tailwindcss";`; `@astrojs/tailwind` is deprecated.
- **Type safety**: `astro check` with `@astrojs/check` and `typescript`; TypeScript 7 native compiler lacks the required programmatic API, so pin `typescript@^5` in `devDependencies`.
- **Testing**: `pnpm exec vitest run` with [vitest.config.ts](references/vitest.config.ts) (`getViteConfig`); browser e2e via [playwright](../playwright/SKILL.md) against `pnpm run preview`.
- **Hygiene & audit**: Knip detects unused dependencies; [knip.json](references/knip.json) excuses task-only and CSS-imported packages; `pnpm audit` scans vulnerabilities.
- **Formatting**: Prettier with `prettier-plugin-astro` formats `.astro`, `.ts`, `.js`, `.css`, `.json`; [prettierignore](references/prettierignore) leaves Markdown, TOML, and YAML to [dprint](../dprint/SKILL.md).
- **Tasks & hooks**: [mise.toml](references/mise.toml) exposes canonical tasks per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).

## 2. Scaffolding Workflow

1. **Create**: `pnpm create astro@latest <slug> --template minimal --install --no-git --no-ai --yes`.
   - Templates: `--template minimal` (clean base), `--template blog` (content layer), `--template starlight` (documentation).
1. **Quality tools**: `pnpm add -D @astrojs/check "typescript@^5" prettier prettier-plugin-astro vitest knip`.
1. **Tailwind CSS v4**: `pnpm add tailwindcss @tailwindcss/vite`, wire the Vite plugin in `astro.config.mjs`, and add `@import "tailwindcss";` to `src/styles/global.css`.
1. **Tasks and hooks**: copy [mise.toml](references/mise.toml), [lefthook.yml](references/lefthook.yml), [knip.json](references/knip.json), [prettierignore](references/prettierignore) as `.prettierignore`, [vitest.config.ts](references/vitest.config.ts), `dprint.json` per [dprint](../dprint/SKILL.md), and [AGENTS.md](references/AGENTS.md).
1. **Validate**: `git init --initial-branch=main`, `mise trust`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`, `mise run build`.
1. **Finish**: `README.md` per [readme-agents](../readme-agents/SKILL.md), then `git add . && git commit -m "chore: initial commit"`.

## 3. Everyday Commands

```bash
astro add react                               # UI framework integrations (react, vue, svelte, preact, solid)
astro add node                                # SSR deployment adapters (node, cloudflare, netlify, vercel)
astro check                                   # diagnostic check for type errors in .astro and .ts files
astro build                                   # production build into dist/
astro preview                                 # local preview server for the built dist/
astro sync                                    # regenerate content collections and module types
```

## 4. Deploy

- **Static (SSG)**: default output; deploy `dist/` to GitHub Pages, Firebase Hosting, Cloudflare Pages, or Netlify.
- **On-demand (SSR)**: add an adapter (`astro add node`, `astro add cloudflare`); deploy Node servers as containers to Cloud Run per [containerize](../containerize/SKILL.md) and [cloud-run](../cloud-run/SKILL.md).

## Gotchas

- **TypeScript 7 incompatibility**: `astro check` fails on TypeScript 7 (`assertCompatibleTypeScript: does not expose the programmatic API`); pin `typescript@^5` in `devDependencies`.
- **`astro check` requires `@astrojs/check`**: without `@astrojs/check` installed, `astro check` prompts interactively and stalls non-interactive gates.
- **Tailwind v4 deprecation**: do NOT use `astro add tailwind` or `@astrojs/tailwind`; install `tailwindcss` and `@tailwindcss/vite` directly.
- **Content Layer API**: modern Astro defines collections in `src/content.config.ts` with `loader` (`glob()` or `file()` from `astro/loaders`), not the legacy `src/content/config.ts`.
- **Prettier and dprint overlap**: `.prettierignore` must ignore `*.md`, `*.yaml`, `*.yml`, and `*.toml` so dprint handles markup and config files without format churn.
- **Knip false positives**: Knip misses `tailwindcss` (referenced only in CSS `@import`) and task runners; keep [knip.json](references/knip.json) rules minimal.

## Official Skills

Upstream ecosystem skill: `astrolicious/agent-skills`. List and install at project scope (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add astrolicious/agent-skills --list
skills add astrolicious/agent-skills --skill astro -y
```

For live documentation retrieval, configure the official Astro Docs remote MCP server (`https://mcp.docs.astro.build/mcp`) via [agent-mcp](../agent-mcp/SKILL.md).

## Documentation

- [Astro](https://astro.build) · [Astro Docs](https://docs.astro.build) · [Content Layer](https://docs.astro.build/en/guides/content-collections/) · [Integrations](https://astro.build/integrations/)
- Companion skills: [typescript-stack](../typescript-stack/SKILL.md), [mise](../mise/SKILL.md), [lefthook](../lefthook/SKILL.md), [dprint](../dprint/SKILL.md), [playwright](../playwright/SKILL.md), [firebase](../firebase/SKILL.md), [cloud-run](../cloud-run/SKILL.md).
