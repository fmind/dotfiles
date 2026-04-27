---
name: install-angular-skills
description: Install Angular's official Agent Skills bundle so the agent gains expert knowledge for Signals, standalone components, control flow, SSR, forms, DI, routing, and accessibility.
---

# Install Angular Skills

Angular publishes the official [`angular/skills`](https://github.com/angular/skills) bundle. It's the canonical source for Angular-specific guidance; this skill explains when and how to install it.

## When to Trigger

- The repo contains `angular.json`, `*.component.ts`, `*.component.html`, `app.config.ts`, or imports from `@angular/*`.
- The user mentions Angular, Signals, standalone components, zoneless, SSR, the new control flow, or `ng generate`.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -i angular`. If installed, skip.

## Install

```bash
# Install the full Angular skills bundle (project scope, default).
npx skills add https://github.com/angular/skills --project
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

2 skills at the time of writing:

- **angular-developer** — Code generation and architectural guidance for Signals, standalone components, forms, DI, routing, SSR, accessibility, and testing.
- **angular-new-app** — Scaffolds new Angular apps with modern CLI best practices.

## Related: `ng` CLI and Angular MCP

The skills assume the Angular CLI is reachable. The MCP server (`ng mcp` or `npx -y @angular/cli mcp`) exposes complementary tools:

```bash
# Default tools: ai_tutor, find_examples, get_best_practices, list_projects,
# onpush_zoneless_migration, search_documentation
ng mcp

# Experimental tools: build, devserver.start/stop/wait_for_build, e2e, modernize, test
ng mcp -E build -E test -E devserver
```

Common CLI idioms the skills will drive:

```bash
ng new my-app --standalone --routing --style=scss
ng generate component features/user-profile
ng build && ng test
```

## After Install

1. Restart the agent so the new skills are picked up by progressive disclosure.
2. The `angular-new-app` skill is the natural entry point when scaffolding; `angular-developer` covers ongoing code generation and refactors.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This skill installs the bundle — it does **not** duplicate Angular docs. Defer Angular-specific guidance to the installed skills.
2. The skills target modern Angular idioms (Signals, standalone, zoneless, new control flow). Don't apply them blindly to legacy NgModule-based codebases without checking the migration path.
3. `npx skills` requires Node.js (already provided by mise).

## Documentation

- [Angular AI hub](https://angular.dev/ai)
- [Agent Skills](https://angular.dev/ai/agent-skills)
- [`angular/skills` repo](https://github.com/angular/skills)
- [Develop with AI (rules files for Cursor, Claude Code, Copilot, Firebase Studio, Gemini CLI)](https://angular.dev/ai/develop-with-ai)
- [Angular MCP server](https://angular.dev/ai/mcp)
- [Angular CLI](https://angular.dev/tools/cli)
