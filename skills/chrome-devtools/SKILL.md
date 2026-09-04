---
name: chrome-devtools
description: "Inspect, debug, and automate Chrome via the Chrome DevTools MCP server and CLI: performance, accessibility, cookies, and memory. Use for live browser debugging and audits."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/chrome-devtools
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Chrome DevTools

Chrome DevTools for agents connects AI coding agents to a live Chrome browser via the Model Context Protocol (MCP) or CLI. It enables automated browser inspection, performance tracing (Core Web Vitals, LCP), accessibility tree evaluation, cookie analysis, network inspection, and memory leak diagnosis. End-to-end browser journeys route through [playwright](../playwright/SKILL.md); modern web standards and native patterns route through [modern-web](../modern-web/SKILL.md).

## 1. Configure MCP Server

Add the `chrome-devtools` MCP server to the current agent harness using [agent-mcp](../agent-mcp/SKILL.md):

```bash
# Project-scoped MCP registration (Antigravity CLI example)
agy mcp add chrome-devtools -- npx -y chrome-devtools-mcp@latest
```

## 2. Debugging and Auditing Workflow

1. **Launch browser session**: Start Chrome with remote debugging enabled, or allow the MCP server to launch and control an isolated browser instance via Puppeteer.
1. **Performance and LCP**: Record traces to identify slow rendering phases, server response latency, render-blocking resources, and Largest Contentful Paint culprits.
1. **Accessibility tree audits**: Inspect computed accessible names, ARIA roles, and color contrast ratios to ensure WCAG compliance alongside [quality-assurance](../quality-assurance/SKILL.md).
1. **Memory and cookies**: Take heap snapshots to identify detached DOM nodes and memory leaks; verify `SameSite`, `Secure`, and `Partitioned` cookie attributes.

## Gotchas

- **Sensitive data exposure**: The MCP server grants full inspection and execution control over the browser session; avoid using personal profiles or sharing sensitive credentials.
- **Headless versus headed**: Automated testing runs headless by default; visual debugging and layout inspections may require headed mode.
- **Complementary to Playwright**: Use Playwright for deterministic functional and E2E regression tests; reserve Chrome DevTools MCP for deep runtime profiling, traces, and live diagnostics.

## Official Skills

Upstream: `ChromeDevTools/chrome-devtools-mcp`. List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add ChromeDevTools/chrome-devtools-mcp --list
skills add ChromeDevTools/chrome-devtools-mcp --skill <name> -y
```

## Documentation

- [Chrome DevTools for agents](https://github.com/ChromeDevTools/chrome-devtools-mcp) · [Chrome DevTools documentation](https://developer.chrome.com/docs/devtools)
- Companion skills: [agent-mcp](../agent-mcp/SKILL.md), [playwright](../playwright/SKILL.md), [modern-web](../modern-web/SKILL.md), [benchmark](../benchmark/SKILL.md), [quality-assurance](../quality-assurance/SKILL.md).
