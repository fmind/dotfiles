---
name: run-lighthouse-audit
description: Run Google Lighthouse audits — single-URL, multi-page, mobile/desktop, with budgets and CI-friendly output (JSON, HTML, JUnit) and Lighthouse CI for trend tracking.
---

# Run Lighthouse Audit

Lighthouse is Google's open-source web auditing tool. It measures performance (Core Web Vitals: LCP, INP, CLS), accessibility (WCAG-aligned), best practices, SEO, and PWA compliance — and surfaces actionable fixes.

This skill drives the **Lighthouse CLI** for one-off audits and **Lighthouse CI (`lhci`)** for trend tracking. The conceptual / fix-it guidance lives in the `web-quality-skills` bundle (install via `install-web-quality-skills`).

## When to Trigger

- The user wants to audit a deployed page (or a local dev server) and surface fixes.
- The user mentions Core Web Vitals, LCP / INP / CLS, "page is slow", performance budgets, or accessibility.
- The user wants Lighthouse CI for regression-tracking across PRs.

## Bootstrap

```bash
# Install via mise (already configured for this user).
mise use -g "npm:lighthouse@latest"
mise use -g "npm:@lhci/cli@latest"   # optional — for Lighthouse CI

# Sanity check.
lighthouse --version
chrome --version           # Lighthouse drives Chrome via CDP
```

On headless servers, ensure Chrome is installed:

```bash
# Debian/Ubuntu.
apt-get install -y chromium                # or google-chrome-stable
# Provide Chrome to Lighthouse via env if needed.
export CHROME_PATH=/usr/bin/chromium
```

## Single-URL Audit

```bash
# HTML report (default), opens in browser when complete.
lighthouse https://example.com --view

# JSON output (CI-friendly).
lighthouse https://example.com \
  --output=json \
  --output-path=./report.json \
  --chrome-flags="--headless --no-sandbox"

# Both formats, written next to each other.
lighthouse https://example.com \
  --output=json --output=html \
  --output-path=./report \
  --chrome-flags="--headless"
```

## Form Factor & Throttling

Lighthouse defaults to **mobile + simulated 4G throttling**. Override:

```bash
# Desktop preset (no throttling, large viewport).
lighthouse https://example.com --preset=desktop --chrome-flags="--headless"

# Mobile, no throttling (closer to dev experience).
lighthouse https://example.com \
  --form-factor=mobile \
  --throttling-method=provided \
  --chrome-flags="--headless"

# Custom CPU & network throttling.
lighthouse https://example.com \
  --throttling.cpuSlowdownMultiplier=4 \
  --throttling.rttMs=150 \
  --throttling.throughputKbps=1638 \
  --chrome-flags="--headless"
```

## Targeted Categories

```bash
# Just performance + a11y (faster, less noise).
lighthouse https://example.com \
  --only-categories=performance,accessibility \
  --output=json --quiet

# All five categories: performance, accessibility, best-practices, seo, pwa.
lighthouse https://example.com --output=json --quiet
```

## Budgets (CI gates)

`budgets.json`:

```json
[
  {
    "path": "/*",
    "timings": [
      { "metric": "interactive", "budget": 4000 },
      { "metric": "largest-contentful-paint", "budget": 2500 },
      { "metric": "cumulative-layout-shift", "budget": 0.1 }
    ],
    "resourceSizes": [
      { "resourceType": "script",   "budget": 250 },
      { "resourceType": "image",    "budget": 200 },
      { "resourceType": "third-party","budget": 100 },
      { "resourceType": "total",    "budget": 800 }
    ],
    "resourceCounts": [
      { "resourceType": "third-party", "budget": 10 }
    ]
  }
]
```

```bash
lighthouse https://example.com \
  --budget-path=./budgets.json \
  --output=json --quiet \
  --chrome-flags="--headless"
```

## Authenticated Pages

```bash
# Pre-set cookies / headers via Chrome flags + extra-headers JSON.
lighthouse https://app.example.com/dashboard \
  --extra-headers='{"Cookie":"session=abc123"}' \
  --chrome-flags="--headless"
```

For complex login flows, use Lighthouse User Flows (programmatic Puppeteer + Lighthouse):

```javascript
// flow.mjs
import { startFlow } from 'lighthouse';
import { launch } from 'puppeteer';

const browser = await launch({ headless: 'new' });
const page = await browser.newPage();
const flow = await startFlow(page, { name: 'Login → Dashboard' });

await flow.navigate('https://app.example.com/login');
await flow.startTimespan({ name: 'login' });
await page.type('#email', 'user@example.com');
await page.type('#password', 'pw');
await page.click('#submit');
await page.waitForNavigation();
await flow.endTimespan();
await flow.snapshot({ name: 'dashboard' });

const report = await flow.generateReport();
// fs.writeFileSync('flow.html', report);
await browser.close();
```

## Multi-page Audits (Lighthouse CI)

Lighthouse CI (`lhci`) is the right tool for routinely auditing multiple URLs and tracking trends.

`lighthouserc.json`:

```json
{
  "ci": {
    "collect": {
      "url": [
        "https://example.com/",
        "https://example.com/blog",
        "https://example.com/pricing"
      ],
      "numberOfRuns": 3,
      "settings": {
        "preset": "desktop",
        "chromeFlags": "--no-sandbox --headless"
      }
    },
    "assert": {
      "preset": "lighthouse:recommended",
      "assertions": {
        "categories:performance":   ["error", { "minScore": 0.9 }],
        "categories:accessibility": ["error", { "minScore": 0.95 }],
        "categories:seo":           ["warn",  { "minScore": 0.9 }]
      }
    },
    "upload": {
      "target": "temporary-public-storage"
    }
  }
}
```

```bash
lhci autorun
```

## CI Integration

```yaml
# .github/workflows/lighthouse.yml
name: lighthouse
on: [pull_request]
jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 22 }
      - run: npm install -g @lhci/cli
      - run: lhci autorun
        env:
          LHCI_GITHUB_APP_TOKEN: ${{ secrets.LHCI_GITHUB_APP_TOKEN }}
```

The Lighthouse CI GitHub App (free) posts results as PR checks.

## Common Workflows

**Audit a deployed page once.**

```bash
lighthouse https://example.com --view
```

**Compare mobile vs desktop quickly.**

```bash
lighthouse https://example.com --form-factor=mobile  --output-path=mobile.html  --chrome-flags="--headless"
lighthouse https://example.com --preset=desktop      --output-path=desktop.html --chrome-flags="--headless"
```

**Run against a local dev server.**

```bash
# In one shell:
npm run dev
# In another:
lighthouse http://localhost:3000 --view --chrome-flags="--headless"
```

**Track regressions across PRs.**

1. Add `lighthouserc.json` (assertions + URLs).
2. Add the GitHub Actions workflow above.
3. Install the [Lighthouse CI GitHub App](https://github.com/apps/lighthouse-ci) on the repo.

## Target Thresholds

The `web-quality-skills` bundle steers towards:

| Metric | Threshold |
|--------|-----------|
| LCP (Largest Contentful Paint) | ≤ 2.5s |
| INP (Interaction to Next Paint) | ≤ 200ms |
| CLS (Cumulative Layout Shift) | ≤ 0.1 |
| Lighthouse Performance score | ≥ 90 |
| Lighthouse Accessibility score | ≥ 95 |

## Important Notes

1. **Lighthouse audits one URL per run** — for SPAs, audit the routes that matter (home, key landing, checkout) rather than just `/`.
2. **`--chrome-flags="--headless"`** is required on CI / headless servers; add `--no-sandbox` if running as root in Docker.
3. **Throttling defaults are aggressive** — mobile + 4G simulation is closer to real-world than what your dev box experiences.
4. **Variance is real** — run 3–5 times and average for trend tracking; one-off scores can swing ±5 points.
5. **Use `lhci autorun`** instead of hand-rolling for repeated runs — it handles aggregation, baseline comparison, and CI uploads.
6. **PWA category is being deprecated** in newer Lighthouse versions — don't gate on it.

## Documentation

- [Lighthouse overview](https://developer.chrome.com/docs/lighthouse/overview)
- [Lighthouse CLI reference](https://github.com/GoogleChrome/lighthouse#using-the-node-cli)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
- [Lighthouse user flows](https://github.com/GoogleChrome/lighthouse/blob/main/docs/user-flows.md)
- [Web Vitals](https://web.dev/articles/vitals)
- [Performance budgets](https://web.dev/articles/use-lighthouse-for-performance-budgets/)
- Companion skills: `install-web-quality-skills`, `install-vercel-skills` (web-design-guidelines).
