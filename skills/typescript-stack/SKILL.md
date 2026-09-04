---
name: typescript-stack
description: "Build TypeScript packages and Node services with pnpm, Biome, tsc, Vitest, and Knip, and route website work to Angular. Use for any TypeScript or Node project."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/typescript-stack
  created: "2026-09-02"
  updated: "2026-09-03"
---

# TypeScript Stack Standard

Canonical TypeScript development: the shared toolchain, the Node package scaffold that Genkit flows, MCP servers, and Firebase Functions sit on, and the architecture choice behind a website. Angular application work belongs to [angular](../angular/SKILL.md); a general HTTP API, CLI, or TUI belongs to [go-stack](../go-stack/SKILL.md) or [python-stack](../python-stack/SKILL.md).

## 1. Core & Quality Stack

- **Runtime and language**: Node LTS (24) with ESM only, TypeScript 7 (the Go-native compiler); `strict` is on by default and `erasableSyntaxOnly` keeps sources runnable by `node src/main.ts` with no build step.
- **Dependencies**: `pnpm` exclusively — `pnpm add`, `pnpm exec`, `pnpm install --frozen-lockfile`; commit `pnpm-lock.yaml`.
- **Tasks and hooks**: [mise.toml](references/mise.toml) exposes the canonical vocabulary per [mise](../mise/SKILL.md); [lefthook.yml](references/lefthook.yml) wires pre-commit and pre-push per [lefthook](../lefthook/SKILL.md).
- **Formatting and linting**: Biome ([biome.json](references/biome.json)) owns TypeScript, JavaScript, JSON, and CSS plus import sorting; dprint ([dprint.json](references/dprint.json)) keeps Markdown, TOML, and YAML per [dprint](../dprint/SKILL.md).
- **Types**: `tsc --noEmit` is the type gate over sources, tests, and configs ([tsconfig.json](references/tsconfig.json)); [tsconfig.build.json](references/tsconfig.build.json) is the only config that emits.
- **Testing**: Vitest with the v8 provider and an 85% coverage gate ([vitest.config.ts](references/vitest.config.ts)); tests are `*.test.ts` files beside the code they cover.
- **Dependency hygiene**: Knip is `check:deps` and starts from auto-detection; [knip.json](references/knip.json) only lists what it provably cannot see.
- **Security**: `pnpm audit --audit-level high` is `check:vuln` and `gitleaks` is `check:leaks`; SAST is opt-in per [opengrep](../opengrep/SKILL.md).
- **Validation and config**: Zod v4 parses every untrusted boundary; `loadConfig()` reads the environment once at startup and fails fast ([config.ts](references/config.ts)).
- **Logging**: `pino` — `pino-pretty` locally, Cloud Logging JSON (`severity`, `message`) in production ([logger.ts](references/logger.ts)) per [observability](../observability/SKILL.md).
- **Publishing**: `publint` as `check:pkg` for anything published to npm; `pnpm dlx @arethetypeswrong/cli --pack .` before the first release.

## 2. Project Scaffolding Workflow

Angular applications skip this section entirely and follow [angular](../angular/SKILL.md).

1. **Information**: define `Slug`, `Description`, and `Holder/Year`.
1. **Bootstrap**: `mkdir <slug> && cd <slug> && pnpm init`, then replace `package.json` with [package.json.template](references/package.json.template) — pin `packageManager` to the installed `pnpm --version` and drop `bin` unless the package ships an executable.
1. **Config files**: [tsconfig.json](references/tsconfig.json), [tsconfig.build.json](references/tsconfig.build.json), [biome.json](references/biome.json), [knip.json](references/knip.json), [vitest.config.ts](references/vitest.config.ts), [mise.toml](references/mise.toml), [lefthook.yml](references/lefthook.yml), [dprint.json](references/dprint.json), `.gitignore` from [gitignore](references/gitignore), `.env.example` from [env.example](references/env.example), `AGENTS.md` from [AGENTS.md](references/AGENTS.md), and `LICENSE` per [project-license](../project-license/SKILL.md).
1. **Sources**: every starter lands in `src/` — [index.ts](references/index.ts), [lib.ts](references/lib.ts), and [lib.test.ts](references/lib.test.ts) for a library; services and executables add [main.ts](references/main.ts), [config.ts](references/config.ts), [logger.ts](references/logger.ts), [config.test.ts](references/config.test.ts), and [logger.test.ts](references/logger.test.ts).
1. **Toolchain**: `mise trust && mise install`, then `pnpm install` (drop `zod`, `pino`, and `pino-pretty` from a library that needs neither); commit the `pnpm-workspace.yaml` pnpm writes.
1. **Validate**: `git init --initial-branch=main`, then `mise run install`, `mise run format`, `mise run check`, `mise run test`, `mise run build`; before the first commit, `check:leaks` scans the working tree.
1. **Finish**: write `README.md` per [readme-agents](../readme-agents/SKILL.md), then `git add . && git commit -m "chore: initial commit"`.

## 3. Project Profiles

- **Library**: `src/index.ts` is the whole public surface and the single `exports` entry; keep runtime dependencies at zero and run `check:pkg` before publishing.
- **Service or executable**: add `src/main.ts` (shebang when `package.json` declares `bin`), typed config, and a logger; run it with `node --watch src/main.ts` in development and `node dist/main.js` in production.
- **Genkit flow, MCP server, Firebase Function**: this scaffold plus the capability skill — [genkit](../genkit/SKILL.md), [mcp-server](../mcp-server/SKILL.md), or [firebase](../firebase/SKILL.md); Functions keep their own `package.json` on the runtime's Node version.
- **Angular website**: [angular](../angular/SKILL.md) owns the toolchain (Angular CLI, angular-eslint, Prettier, its own pinned TypeScript); only the conventions in §1 that Angular does not override apply.

## 4. Website Architecture

Choose the smallest architecture before scaffolding; a Firebase-backed client-side app is already full stack, and SSR is a rendering choice rather than a prerequisite for backend features.

| Need                                             | Default                               | Add                                                             |
| ------------------------------------------------ | ------------------------------------- | --------------------------------------------------------------- |
| Authenticated or highly interactive UI           | Client-side rendering (CSR)           | [angular](../angular/SKILL.md) and static hosting               |
| Public pages with stable content                 | Static generation (SSG) per route     | Angular server routing and static hosting                       |
| Public pages with request-time data              | Server-side rendering (SSR) per route | Angular server routing and Firebase App Hosting                 |
| Authentication, data, files, or server functions | Firebase                              | [firebase](../firebase/SKILL.md) and local emulators            |
| Generative AI inside the product                 | Genkit behind a server boundary       | [genkit](../genkit/SKILL.md), usually deployed through Firebase |

- **Shape by feature**: colocate each route's components, state, data access, and tests; keep composition roots thin and lazy-load routes instead of creating generic `shared`, `core`, or `utils` layers pre-emptively.
- **Web quality**: semantic HTML, keyboard and focus behavior, accessible loading and error states, bundle budgets, and per-route rendering chosen from measured SEO and performance needs.
- **Test by risk**: units and components in Vitest, dependency hygiene in Knip, a small set of critical journeys in [playwright](../playwright/SKILL.md), Firebase emulator tests and Genkit evaluations only once those layers exist.
- **Deploy separately**: deployment requires explicit authorization and follows [firebase](../firebase/SKILL.md) for Hosting or App Hosting, or [cloud-run](../cloud-run/SKILL.md) for a container.

## Gotchas

- **TypeScript 7 changed defaults**: `strict` is on, `module` defaults to `esnext`, and `types` defaults to `[]` — list `["node"]` explicitly or every ambient `@types` package disappears. The programmatic API only stabilizes in 7.1, so tools embedding the compiler may still need TypeScript 6.
- **Angular pins its own TypeScript**: Angular 22 peers `typescript >=6.0 <6.1`, so an Angular repository is not on TypeScript 7; let `ng update` move it and never share one `tsconfig` across both.
- **pnpm 11 guards the supply chain**: `minimumReleaseAge` defaults to one day, so a just-published version resolves only after `pnpm-workspace.yaml` lists it under `minimumReleaseAgeExclude`, and dependency build scripts are blocked until `pnpm approve-builds` records them under `allowBuilds`. pnpm writes that file even in a single-package repository — commit it.
- **Global tools are not project pins**: the global `tsc`, `biome`, and `knip` from mise are conveniences; repository tasks always use `pnpm exec` so hooks and CI run the lockfile's versions.
- **One formatter per file**: Biome owns `**/*.json`, so `dprint.json` must exclude it; Angular projects replace Biome with the Angular skill's Prettier split because Biome does not format Angular templates.
- **Biome's type-aware rules cost 20x**: enabling the `types` domain (`noFloatingPromises`, `noMisusedPromises`) starts the project scanner — measured at ~14s against ~0.6s for the `project` and `test` domains and ~0.8s for a full `tsc --noEmit`. Leave it off unless unhandled rejections are a real risk, and enable the rules explicitly: they ship in `nursery` at `info` severity, so the domain alone never fails `check`.
- **Node runs TypeScript, not all of it**: type stripping rejects `enum`, `namespace`, and parameter properties — `erasableSyntaxOnly` catches them at type-check time. Reach for `tsx` only when a project needs `tsconfig` path aliases, decorators, or CJS interop.
- **Import the `.ts` extension**: `allowImportingTsExtensions` plus `rewriteRelativeImportExtensions` lets the same source run under `node` and emit `./x.js`; writing `./x.js` in source breaks running from `src/`.
- **Knip cannot see mise tasks**: a tool invoked only by `mise run` (Biome, publint) or referenced by string (`pino-pretty` as a pino transport) is reported unused — add it to `ignoreDependencies`, never a `package.json` script that duplicates the task.
- **The browser is untrusted**: Firebase web configuration ships to clients, but service-account credentials, model keys, and privileged operations stay server-side; Admin SDK calls bypass Security Rules, so server handlers authorize every operation themselves, and App Check is defense in depth — monitor legitimate traffic before enforcing, because enforcement rejects unverified clients.

## Official Skills

Use the upstream bundle owned by the selected capability. Each companion skill lists its bundle with `--list` before installing a reviewed skill: [angular](../angular/SKILL.md), [firebase](../firebase/SKILL.md), and [genkit](../genkit/SKILL.md).

## Documentation

- [TypeScript](https://www.typescriptlang.org/docs/) · [Node](https://nodejs.org/docs/latest/api/) · [pnpm](https://pnpm.io) · [Biome](https://biomejs.dev) · [Vitest](https://vitest.dev) · [Knip](https://knip.dev) · [Zod](https://zod.dev) · [pino](https://getpino.io)
- Companion skills: [angular](../angular/SKILL.md), [firebase](../firebase/SKILL.md), [genkit](../genkit/SKILL.md), [mcp-server](../mcp-server/SKILL.md), [playwright](../playwright/SKILL.md), [containerize](../containerize/SKILL.md), [github-actions](../github-actions/SKILL.md), [secure](../secure/SKILL.md).
