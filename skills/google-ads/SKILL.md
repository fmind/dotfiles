---
name: google-ads
description: Map Google Ads work (Ads API, Data Manager API, Mobile Ads SDK, IMA SDK) to the official google/skills ads catalog and install what a task needs. Use for any Google Ads task.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/google-ads
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Google Ads

Google's advertising developer surface has four families: the Google Ads API (campaigns, reporting, an official MCP server), the Data Manager API (audiences and conversions), the Google Mobile Ads SDK (AdMob and Ad Manager in apps), and the IMA SDK (video ads). Marketing strategy is out of scope; this skill routes the engineering task to the `ads` group of the official catalog.

## Gotchas

- **Credentials**: the Ads API needs a developer token plus an OAuth client; keep both in environment variables or Secret Manager.
- **Money moves**: use a test account and `validate_only` requests before touching a live account; every mutation can spend budget.
- **MCP**: the Ads MCP server runs locally with the same credentials; configure it per [agent-mcp](../agent-mcp/SKILL.md).

## Official Skills

Upstream: `google/skills` (`skills/ads` group; pick by family: API quickstart or diagnostics, Data Manager ingestion, mobile SDK formats and migration, IMA client-side or DAI). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add https://github.com/google/skills/tree/main/skills/ads --list
skills add https://github.com/google/skills/tree/main/skills/ads --skill <name> -y
```

## Documentation

- [Google Ads API](https://developers.google.com/google-ads/api/docs/start) · [AdMob](https://developers.google.com/admob) · [IMA SDK](https://developers.google.com/interactive-media-ads)
- Companion skills: [google-analytics](../google-analytics/SKILL.md), [google-developer](../google-developer/SKILL.md), [gcloud](../gcloud/SKILL.md), [agent-mcp](../agent-mcp/SKILL.md).
