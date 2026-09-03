---
name: angular
description: "Build Angular apps with the Angular CLI: standalone components, signals, zoneless, Vitest, angular-eslint, mise tasks. Use for Angular apps, routing, forms, SSR."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/angular
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Angular Standard

Angular owns its toolchain: the CLI builds, serves, tests, generates, and updates; this skill wraps it in the shared task vocabulary. Shared TypeScript conventions come from [typescript-stack](../typescript-stack/SKILL.md); backends and hosting from [firebase](../firebase/SKILL.md).

## 1. Core Stack

- **CLI**: `ng` from mise (`npm:@angular/cli`); every project pins its own `@angular/cli` and `@angular/build` (esbuild application builder) in `package.json`, installed with `pnpm`.
- **Architecture**: standalone components, signals for state (`signal`, `computed`, `linkedSignal`, `resource`, `httpResource`), zoneless change detection (`--zoneless`), Signal Forms for new forms, lazy routes.
- **Tests**: `ng test` runs Vitest with jsdom; coverage needs `@vitest/coverage-v8`; e2e through `ng add playwright-ng-schematics` when a journey needs a browser (browsers via `playwright install chromium`).
- **Dependency hygiene**: Knip detects unused files, exports, and dependencies; its Angular plugin activates automatically from `@angular/cli`, and [knip.json](references/knip.json) only excuses the two things it cannot see in a fresh workspace.
- **Lint and format**: `ng add angular-eslint` gives `ng lint` for TypeScript and templates (accessibility rules included); Prettier formats TypeScript, templates, CSS, and JSON; Biome does not understand Angular templates.
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).
- **Formatter split**: [prettierignore](references/prettierignore) leaves Markdown, TOML, and YAML to dprint; `dprint.json` excludes `**/*.json`, `dist`, and `.angular`.

## 2. Scaffolding Workflow

1. **Create**: `ng new <slug> --package-manager pnpm --style css --zoneless --ssr false --skip-git --defaults --ai-config none`.
   - Options: `--style tailwind` wires Tailwind at creation, `--ssr true` for SEO or server rendering, `--prefix <prefix>` sets the selector prefix.
1. **Quality tools**: `cd <slug> && ng add angular-eslint --skip-confirmation`, then `pnpm add -D @vitest/coverage-v8 knip`.
1. **Tasks and hooks**: copy [mise.toml](references/mise.toml), [lefthook.yml](references/lefthook.yml), [knip.json](references/knip.json), [prettierignore](references/prettierignore) as `.prettierignore`, `dprint.json` per [dprint](../dprint/SKILL.md) with the excludes above, and `AGENTS.md` from [AGENTS.md](references/AGENTS.md).
1. **Validate**: `git init --initial-branch=main`, `mise trust`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`, `mise run build`.
1. **Finish**: `README.md` per [readme-agents](../readme-agents/SKILL.md), then `git add . && git commit -m "chore: initial commit"`.

## 3. Everyday Commands

```bash
ng generate component features/user-card     # also: service, guard, pipe, directive, interceptor
ng add @angular/material                     # Angular libraries always through ng add, never pnpm add
ng add tailwindcss                           # Tailwind v4 wired into the build (or `ng new --style tailwind`)
ng add @angular/fire                         # Firebase SDK and deploy target (firebase skill; see Gotchas)
ng update @angular/core @angular/cli         # framework upgrades, one major at a time (upgrade-tools skill)
ng build --configuration development         # fast debug bundle; `mise run build` is production
```

## 4. Deploy

- **Static SPA**: `mise run build` then Firebase Hosting or any static host per [firebase](../firebase/SKILL.md).
- **SSR**: `--ssr true` projects deploy to Firebase App Hosting, or as a Node container to Cloud Run per [containerize](../containerize/SKILL.md) and [cloud-run](../cloud-run/SKILL.md).

## Gotchas

- **TypeScript is pinned by Angular**: `ng new` installs the TypeScript major Angular supports, not the newest; let `ng update` move it.
- **`@angular/fire` lags Angular majors**: check `pnpm view @angular/fire peerDependencies` before `ng add`; when its `@angular/core` peer is behind, use the plain `firebase` SDK until it catches up.
- **`ng test` watches by default**: tasks and CI pass `--watch=false`.
- **`--ai-config`** accepts `claude-code`, `gemini-cli`, `open-ai-codex`, `cursor`, `vscode`, or `none`; use `none` and the shared `AGENTS.md` from [agent-project](../agent-project/SKILL.md).
- **Prettier and dprint overlap**: `.prettierignore` must list `*.md`, `*.yaml`, `*.yml`, `*.toml`, or both formatters rewrite the same files.
- **Knip cannot see mise tasks**: a fresh `ng new` workspace fails `check:deps` on `prettier` (invoked only by `mise run`) and `@angular/forms` (installed by the CLI, unused until the first form), which is why [knip.json](references/knip.json) ignores exactly those two; add entry points or ignores only for further gaps it demonstrably cannot see.
- **Zoneless tests**: `fixture.detectChanges()` is not enough; `await fixture.whenStable()` after state changes.

## Official Skills

Upstream: `angular/skills`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add angular/skills --list
skills add angular/skills --skill <name> -y
```

## Documentation

- [Angular](https://angular.dev) · [Angular CLI](https://angular.dev/tools/cli) · [angular-eslint](https://github.com/angular-eslint/angular-eslint) · [Vitest](https://vitest.dev)
- Companion skills: [typescript-stack](../typescript-stack/SKILL.md), [firebase](../firebase/SKILL.md), [product-design-review](../product-design-review/SKILL.md) (UI audits).
