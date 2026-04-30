---
name: install-web-quality-skills
description: Install web-quality-skills (Addy Osmani / Chrome) for Lighthouse-style audits — Core Web Vitals, performance, a11y, SEO, best practices.
---

# Install Web Quality Skills (Lighthouse / Core Web Vitals)

[`addyosmani/web-quality-skills`](https://github.com/addyosmani/web-quality-skills) is a stack-agnostic skill bundle covering the same pillars Lighthouse audits — performance, Core Web Vitals (LCP / INP / CLS), accessibility (WCAG), SEO, and modern best practices. Authored by Addy Osmani (Google Chrome team), MIT-licensed.

This skill explains when and how to install the bundle; the actual audit / optimization expertise lives in the installed skills.

## When to Trigger

- The user is working on a website / web app and asks about performance, Core Web Vitals, accessibility, SEO, or wants a Lighthouse audit.
- The repo contains `next.config.js`, `nuxt.config.ts`, `astro.config.mjs`, `vite.config.*`, or other web framework markers.
- Verify first: `ls .agents/skills/ ~/.gemini/skills/ 2>/dev/null | grep -E 'web-quality|core-web-vitals|performance|accessibility|seo'`. If installed, skip.

## Install

```bash
npx skills add addyosmani/web-quality-skills
```

`npx skills` writes to `.agents/skills/` for project scope (default — repo-pinned, commits with the codebase) or to `~/.gemini/skills/` for global scope when prompted.

## What Gets Installed

| Skill              | Covers                                                     |
|--------------------|------------------------------------------------------------|
| `web-quality-audit`| Comprehensive review across all pillars (entry point)      |
| `performance`      | Loading speed, resource budgets, bundle hygiene            |
| `core-web-vitals`  | Targeted LCP / INP / CLS optimizations                     |
| `accessibility`    | WCAG compliance, screen-reader support                     |
| `seo`              | Search-engine optimization, structured data                |
| `best-practices`   | Security headers, modern APIs, code quality                |

Target thresholds the skills steer towards: LCP ≤ 2.5s, INP ≤ 200ms, CLS ≤ 0.1, Lighthouse Performance ≥ 90.

The skills include framework-specific guidance for React/Next, Vue/Nuxt, Astro, Svelte, Angular, and plain HTML.

## Related: `lighthouse` CLI

The skills assume Lighthouse can be run directly. The CLI is installed via mise (`npm:lighthouse`):

```bash
# Headless audit, JSON output.
lighthouse https://example.com --output=json --output-path=./report.json --chrome-flags="--headless"

# HTML report opened in a browser.
lighthouse https://example.com --view

# Mobile-only, throttled (default), specific categories.
lighthouse https://example.com \
  --only-categories=performance,accessibility \
  --form-factor=mobile

# CI-friendly: budgets + assertions.
lighthouse https://example.com \
  --budget-path=./budgets.json \
  --output=json --quiet
```

For multi-page / regression tracking, use `lhci` (Lighthouse CI) — out of scope for this skill, but the bundle's `performance` skill points to it.

## After Install

1. Restart the agent so progressive disclosure picks up the new descriptions.
2. The `web-quality-audit` skill is the natural entry point — call it for a holistic review, then drill into pillar-specific skills as needed.
3. Project-scope installs (`.agents/skills/`) commit naturally with the repo. Global-scope installs are machine-local — `chezmoi add ~/.gemini/skills/<slug>` to track them.

## Important Notes

1. This is a community / personal repo (MIT, by a Chrome team member) — not an "official Google" bundle. Treat it as high-quality but unaffiliated.
2. Skills are stack-agnostic but framework-aware; expect framework-specific suggestions (e.g. `next/image` for Next.js, `nuxt/image` for Nuxt).
3. Lighthouse CLI runs Chrome under the hood — it must be installed and reachable. On headless servers, ensure required system libs.
4. Lighthouse audits a single URL per run; for SPAs, audit the routes that matter (home, key landing pages, checkout) rather than just `/`.

## Documentation

- [`addyosmani/web-quality-skills`](https://github.com/addyosmani/web-quality-skills)
- [Lighthouse docs](https://developer.chrome.com/docs/lighthouse/overview)
- [web.dev — Core Web Vitals](https://web.dev/articles/vitals)
- [`npx skills` CLI](https://github.com/vercel-labs/skills)
