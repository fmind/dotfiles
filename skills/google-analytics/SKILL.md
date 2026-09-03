---
name: google-analytics
description: Map Google Analytics API work (Admin API for properties, Data API for reports) to the official google/skills analytics catalog and install what a task needs. Use for any Google Analytics task.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/google-analytics
  created: "2026-09-03"
  updated: "2026-09-03"
---

# Google Analytics

Google Analytics (GA4) exposes two developer APIs: the Admin API manages accounts, properties, data streams, custom dimensions, and key events; the Data API runs reports against a property. Both are Google Cloud APIs enabled on a pinned project per [gcloud](../gcloud/SKILL.md). Dashboards and the web UI are out of scope; this skill routes API work to the `analytics` group of the official catalog.

## Gotchas

- **Identifiers**: API calls take the numeric property (`properties/<id>`), not the `G-...` measurement ID from the tag.
- **Quotas**: limits apply per property and per project; batch report requests and cache results locally (see [duckdb](../duckdb/SKILL.md)) instead of re-querying.
- **Personal data**: reports and exports are personal data by default; keep raw exports out of git and follow the project's retention rule.

## Official Skills

Upstream: `google/skills` (`skills/analytics` group; Admin-type skills cover configuration, Data-type skills cover reporting). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add https://github.com/google/skills/tree/main/skills/analytics --list
skills add https://github.com/google/skills/tree/main/skills/analytics --skill <name> -y
```

## Documentation

- [Analytics Admin API](https://developers.google.com/analytics/devguides/config/admin/v1) · [Analytics Data API](https://developers.google.com/analytics/devguides/reporting/data/v1)
- Companion skills: [gcloud](../gcloud/SKILL.md), [duckdb](../duckdb/SKILL.md), [google-ads](../google-ads/SKILL.md), [google-developer](../google-developer/SKILL.md).
