# AGENTS.md (Project)

Context and rules for AI agents working in this repository. Humans should start with `README.md`.

## Project overview

- **Name**: <slug>
- **Description**: <1-2 sentences on what this application does.>
- **Stack**: Astro 7 on TypeScript, built with Astro, Tailwind CSS v4, and `pnpm`.

## Setup & core commands

All work goes through `mise` (see `mise.toml`); git hooks and CI call the same tasks.

- Install: `mise run install` — `pnpm install --frozen-lockfile` and git hooks.
- Format: `mise run format` — Prettier (Astro, TypeScript, CSS, JSON) and dprint (Markdown, TOML, YAML).
- Check: `mise run check` — `astro check`, Knip dependency hygiene, format checks, `pnpm audit`, gitleaks.
- Test: `mise run test` — `vitest run` once with coverage or passWithNoTests.
- Build: `mise run build` — `astro build` production bundle into `dist/`.
- Watch: `mise run watch` — `astro dev` with live reload.

## Definition of done

A change is complete only when, locally, `mise run format` is clean, `mise run check` reports no findings, `mise run test` is green, `mise run build` succeeds, and new or changed behavior has a test. Fix root causes — never weaken an assertion, skip a test, or disable a lint rule to force a green result.

## Conventions & idioms

- **Framework commands**: `astro add <integration>` for official integrations; `pnpm exec astro check` for type diagnostics.
- **Islands**: zero JavaScript by default; use client directives (`client:load`, `client:idle`, `client:visible`) only where interactivity is required; use `server:defer` for deferred server islands.
- **Styling**: Tailwind CSS v4 via `@tailwindcss/vite` and `@import "tailwindcss";`; do not add `@astrojs/tailwind` or `tailwind.config.js`.
- **Content**: Content Layer API with `defineCollection` and loaders in `src/content.config.ts`.
- **TypeScript**: TypeScript 5.x pinned in devDependencies (`typescript@^5`) for `@astrojs/check` compatibility.
- **Dependencies**: `pnpm exec knip` must stay clean; configure exceptions only for generated or dynamically referenced entry points Knip cannot discover.
- **Accessibility**: semantic HTML first, accessible forms and navigation; keep interactive elements keyboard-reachable.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`); no attribution in commit messages.

## Repository layout

- `src/pages/` — file-based routing; `index.astro` is the home page.
- `src/components/` — reusable Astro and framework components.
- `src/layouts/` — page layouts and shell templates.
- `src/styles/` — global CSS styles (`global.css`).
- `src/content.config.ts` — Content Layer collections definition and schemas.
- `public/` — static assets copied as-is to `dist/`.
- `astro.config.mjs`, `tsconfig.json`, `.prettierrc` — Astro toolchain; `mise.toml`, `lefthook.yml`, `dprint.json` — task and hook wiring.
