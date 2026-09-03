---
name: playwright
description: Drive browsers with Playwright for end-to-end tests, screenshots, traces, and codegen, and install the official Playwright skills. Use for browser automation or e2e testing.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/playwright
  created: "2026-09-02"
  updated: "2026-09-03"
---

# Playwright

Use Playwright for browser automation and end-to-end tests. Test strategy belongs to [quality-assurance](../quality-assurance/SKILL.md) and interface critique to [product-design-review](../product-design-review/SKILL.md); this skill owns the tool: browsers, commands, and the official skills.

## Workflow

1. **Pin the browser**: `playwright install chromium` (plus `playwright install-deps` on a fresh Linux host); the mise binary is the global CLI, a project pins `@playwright/test` with pnpm per [typescript-stack](../typescript-stack/SKILL.md).
1. **Explore and record**: `playwright codegen <url>` records actions into a test; `playwright screenshot <url> <file>` and `playwright pdf <url> <file>` produce review evidence.
1. **Run tests**: `playwright test` in a project with `playwright.config.ts`; add `--trace on` when a failure needs context, then `playwright show-trace <trace.zip>` or `playwright show-report`.
1. **Expose the browser to agents**: `playwright mcp` serves the browser over MCP per [agent-mcp](../agent-mcp/SKILL.md); `playwright init-agents --loop <claude|codex|copilot|opencode>` seeds planner, generator, and healer agents.
1. **Verify**: a green `playwright test` run plus the artifact (screenshot, trace, or report) the task asked for.

## Gotchas

- **Authority**: a test request does not authorize reusing a logged-in browser, synchronizing cookies, entering passwords or MFA, creating accounts, bypassing CAPTCHA, accepting legal terms, making purchases, or paying for cloud browsers or tunnels; stop and ask.
- **Browser cache**: binaries live in `~/.cache/ms-playwright`; `playwright uninstall` frees them (see [reclaim-disk](../reclaim-disk/SKILL.md)).
- **Version skew**: browsers match the Playwright version that installed them; rerun `playwright install` after an upgrade.
- **Headless by default**: pass `--headed` to watch a run; keep CI headless.

## Official Skills

Upstream: `microsoft/playwright`. The CLI installs the skills bundled with the installed Playwright version into the project's `.agents/skills`; review the snapshot before trusting it (see [agent-skills](../agent-skills/SKILL.md)):

```bash
playwright init-skills --loop agents
```

## Documentation

- [Playwright](https://playwright.dev/docs/intro) · [playwright-cli](https://github.com/microsoft/playwright-cli)
- Accessibility and performance evidence: `ChromeDevTools/chrome-devtools-mcp` ships skills for its MCP server (`skills add ChromeDevTools/chrome-devtools-mcp --list`); `lighthouse <url> --output json` stays the one-shot audit.
- Companion skills: [quality-assurance](../quality-assurance/SKILL.md), [product-design-review](../product-design-review/SKILL.md), [angular](../angular/SKILL.md), [benchmark](../benchmark/SKILL.md) (load, not browser, testing).
