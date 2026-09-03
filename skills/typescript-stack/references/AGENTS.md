# AGENTS.md (Project)

Context and rules for AI agents working in this repository. Humans should start with `README.md`.

## Project overview

- **Name**: <slug>
- **Description**: <1-2 sentences on what this package does.>
- **Stack**: TypeScript 7 on Node 24 ESM, managed with `pnpm`; Biome formats and lints, `tsc` type-checks and emits, Vitest tests.

## Setup & core commands

All work goes through `mise` (see `mise.toml`); git hooks and CI call the same tasks.

- Install: `mise run install` — `pnpm install --frozen-lockfile` and git hooks.
- Format: `mise run format` — Biome (TypeScript, JSON, CSS, imports) and dprint (Markdown, TOML, YAML).
- Check: `mise run check` — Biome lint, `tsc --noEmit`, Knip dependency hygiene, format checks, `pnpm audit`, gitleaks.
- Test: `mise run test` — Vitest once with an 85% coverage gate.
- Build: `mise run build` — `tsc --project tsconfig.build.json` into `dist/`.
- Watch: `mise run watch` — Vitest in watch mode; run a service with `node --watch src/main.ts`.

## Definition of done

A change is complete only when, locally, `mise run format` is clean, `mise run check` reports no findings, `mise run test` is green, `mise run build` succeeds, and new or changed behavior has a test. Fix root causes — never weaken an assertion, skip a test, loosen a type, or suppress a lint rule to force a green result.

## Conventions & idioms

- **Modules**: ESM only (`"type": "module"`); relative imports carry the real `.ts` extension and `rewriteRelativeImportExtensions` emits `.js`.
- **Erasable syntax only**: no `enum`, `namespace`, or parameter properties, so `node src/main.ts` runs the sources unchanged.
- **Types**: no `any` and no assertion used to bypass validation; parse untrusted input (environment, HTTP, storage, model output) with Zod at the boundary and pass typed values inward.
- **Errors**: `throw new Error(...)` with `{ cause }` for context; never swallow a rejection, and `await`, `return`, or `void` every promise.
- **Config**: `loadConfig()` parses the environment once at startup and fails fast; no `process.env` reads scattered through the code.
- **Logging**: `pino` — pretty in development, Cloud Logging JSON (`severity`, `message`) in production; no `console.log` outside a CLI's own output.
- **Tests**: `*.test.ts` beside the code it covers; deterministic and offline by default.
- **Dependencies**: `pnpm exec knip` must stay clean; add a `knip.json` exception only for a tool invoked by `mise` or referenced by string.
- **Commits**: Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`); no attribution in commit messages.

## Repository layout

- `src/index.ts` — public surface of the package, the single `exports` entry.
- `src/main.ts` — executable entry point (services and CLIs); starts with a shebang when `package.json` declares `bin`.
- `src/config.ts` — Zod schema over the environment; `src/logger.ts` — pino logger factory.
- `src/**/*.test.ts` — unit tests beside their sources.
- `tsconfig.json` type-checks (`noEmit`); `tsconfig.build.json` emits `dist/`; `biome.json`, `knip.json`, `vitest.config.ts` — tooling.
- `mise.toml` — tasks and pinned tools; `lefthook.yml` — git hooks; `dprint.json` — Markdown, TOML, YAML formatter.
- `.env` / `.env.example` — environment configuration (never commit secrets).
