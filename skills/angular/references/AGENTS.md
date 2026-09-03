# AGENTS.md (Project)

Context and rules for AI agents working in this repository. Humans should start with `README.md`.

## Project overview

- **Name**: <slug>
- **Description**: <1-2 sentences on what this application does.>
- **Stack**: Angular 22 (standalone, signals, zoneless) on TypeScript, built with the Angular CLI and `pnpm`.

## Setup & core commands

All work goes through `mise` (see `mise.toml`); git hooks and CI call the same tasks.

- Install: `mise run install` — `pnpm install --frozen-lockfile` and git hooks.
- Format: `mise run format` — Prettier (TypeScript, templates, CSS, JSON) and dprint (Markdown, TOML, YAML).
- Check: `mise run check` — `ng lint` (angular-eslint), Knip dependency hygiene, format checks, `pnpm audit`, gitleaks.
- Test: `mise run test` — `ng test` once with coverage (Vitest, jsdom).
- Build: `mise run build` — `ng build` production bundle into `dist/`.
- Watch: `mise run watch` — `ng serve` with live reload.

## Definition of done

A change is complete only when, locally, `mise run format` is clean, `mise run check` reports no findings, `mise run test` is green, `mise run build` succeeds, and new or changed behavior has a test. Fix root causes — never weaken an assertion, skip a test, or disable a lint rule to force a green result.

## Conventions & idioms

- **Generate, do not hand-write**: `ng generate component|service|guard|pipe <name>`; `ng add <package>` for Angular libraries (Material, Tailwind, AngularFire).
- **Reactivity**: signals (`signal`, `computed`, `linkedSignal`, `resource`, `httpResource`) over manual subscriptions; no `zone.js`, so tests use `await fixture.whenStable()` rather than `detectChanges()`.
- **Forms**: Signal Forms for new forms; reactive forms only where they already exist.
- **Naming**: Angular v20+ style guide — file names by intent (`user-card.ts`), no `.component` suffix, selectors prefixed by the project prefix.
- **HTTP**: `provideHttpClient(withFetch())`, typed responses parsed at the boundary, errors surfaced through the UI state, never swallowed.
- **Dependencies**: `pnpm exec knip` must stay clean; configure exceptions only for generated or dynamically referenced entry points Knip cannot discover.
- **Accessibility**: semantic elements first, ARIA only when needed, every interactive element keyboard-reachable; lint enforces the template rules.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`); no attribution in commit messages.

## Repository layout

- `src/app/app.ts`, `app.html`, `app.css` — root component; `app.config.ts` providers; `app.routes.ts` routes.
- `src/app/<feature>/` — one folder per feature with its components, services, and specs beside the code.
- `public/` — static assets copied as-is.
- `angular.json`, `tsconfig*.json`, `eslint.config.js`, `.prettierrc` — Angular CLI toolchain; `mise.toml`, `lefthook.yml`, `dprint.json` — task and hook wiring.
