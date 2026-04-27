---
name: install-design-md
description: Install the official Google Labs `design.md` skills bundle so the agent gains expert knowledge for typed service contracts, TDD discipline, agent-friendly CLI design, and Ink-based terminal UIs.
---

# Install design.md Skills

Google Labs publishes the official [`google-labs-code/design.md`](https://github.com/google-labs-code/design.md) repo, which is both the home of the [DESIGN.md](https://github.com/google-labs-code/design.md) format spec **and** a small bundle of agent-oriented engineering skills. This skill explains when and how to install the skill bundle.

> Not to be confused with the **`design-md`** skill inside `install-stitch-skills` — that one generates `DESIGN.md` files for visual design systems. This bundle is about TypeScript service architecture, TDD, CLI ergonomics, and terminal UIs.

## When to Trigger

- The user wants disciplined TDD (red-green-refactor) on a TypeScript / Node.js project.
- The user is building or evaluating a CLI intended to be driven by AI agents.
- The user is architecting type-safe TypeScript services with strict input parsing and exhaustive error handling.
- The user is building interactive terminal UIs from JSON specs (Ink renderer for `json-render`).
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'tdd-red-green-refactor|typed-service-contracts|agent-dx-cli-scale|^ink$'`. If installed, skip.

## Install

```bash
# List available skills.
npx skills add google-labs-code/design.md --list

# Install one (repeat per skill — project scope by default).
npx skills add google-labs-code/design.md --skill tdd-red-green-refactor --project
npx skills add google-labs-code/design.md --skill typed-service-contracts --project
npx skills add google-labs-code/design.md --skill agent-dx-cli-scale --project
npx skills add google-labs-code/design.md --skill ink --project
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted. Cross-tool: Claude Code, Gemini CLI, Cursor, and Antigravity all read `.agents/skills/`.

## What Gets Installed

4 skills at the time of writing:

- **tdd-red-green-refactor** — Enforces a disciplined Red-Green-Refactor TDD workflow in TypeScript / Node.js: write a failing test, ship the minimal implementation, then refactor under the type system.
- **typed-service-contracts** — "Spec & Handler" architecture standard for robust, type-safe TypeScript services with strict input parsing and exhaustive error handling.
- **agent-dx-cli-scale** — Scoring scale (0–21 across 7 axes — machine-readable output, safety rails, etc.) for evaluating how agent-friendly a CLI is, based on the "Rewrite Your CLI for AI Agents" principles.
- **ink** — Terminal renderer for `json-render` (`@json-render/ink`) — turns JSON specs into interactive terminal component trees with data binding and event handling.

## Related: the design.md CLI itself

The same repo ships the `@google/design.md` CLI, which is a separate concern from the skill bundle:

```bash
npx @google/design.md lint DESIGN.md             # validate structural correctness, WCAG contrast
npx @google/design.md diff DESIGN.md DESIGN-v2.md # detect token-level regressions
npx @google/design.md export --format tailwind DESIGN.md   # tokens → Tailwind theme
npx @google/design.md export --format dtcg DESIGN.md       # tokens → W3C DTCG tokens.json
npx @google/design.md spec --rules               # print the format spec for context priming
```

Use the CLI when you have an actual `DESIGN.md` file to lint / diff / export. Use the skill bundle (this skill) when you want the agent to internalize the engineering practices listed above.

## After Install

1. Restart the agent so the new skill descriptions are picked up by progressive disclosure.
2. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.
3. `tdd-red-green-refactor` and `typed-service-contracts` are the most generally useful — pull them first on most TypeScript projects. `agent-dx-cli-scale` is for CLI-design audits; `ink` is only relevant if you're building a terminal UI.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate the upstream skill content. Defer to the installed skills.
2. The bundled skills are TypeScript-flavoured. They still convey transferable principles in other languages, but the worked examples assume TS / Node.
3. Don't install all 4 reflexively — pick the ones that match the project. The bundle is small enough that over-installing is cheap, but over-triggering on irrelevant work is the real cost.
4. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [`google-labs-code/design.md` repo](https://github.com/google-labs-code/design.md)
- [`design.md` on skills.sh](https://skills.sh/google-labs-code/design.md)
- [Rewrite Your CLI for AI Agents](https://github.com/google-labs-code/design.md) — the principles behind `agent-dx-cli-scale`
- [Ink (terminal UI library)](https://github.com/vadimdemedes/ink)
