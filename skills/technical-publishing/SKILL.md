---
name: technical-publishing
description: Publish technical articles through package review, canonical web export, and channel copy. Use when drafting, checking, verifying, or publishing an article.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/technical-publishing
  created: "2026-08-30"
  updated: "2026-09-03"
---

# Publish Technical Articles

Publish technical articles from package sources through review, canonical web export, and channel copy. Third-party documentation sites and static websites use [hugo](../hugo/SKILL.md), software releases use [release](../release/SKILL.md), and standalone documents use [typst](../typst/SKILL.md).

```bash
pub init article <slug>                          # scaffold package directories and draft
pub check <package> --publish-ready              # offline gate; readiness, not authorization
pub publish <package> --site <site-directory>    # export to canonical site; --dry-run previews
```

## Workflow

1. **Package pipeline**: work inside the article package layout (`draft.txt -> article.md -> canonical site -> posts/`); `draft.txt` contains raw notes that agents never edit; see [Package layout](references/packages.md).
1. **Voice and boundaries**: adhere to the editorial voice in [Voice and identity](references/voice.md); agree the register before drafting; never invent anecdotes, clients, or metrics; keep published articles immutable.
1. **Visuals**: explanatory diagrams follow [fmind-visuals](../fmind-visuals/SKILL.md) and [d2](../d2/SKILL.md) using the light-surface theme; store `.d2` sources beside rendered PNGs.
1. **Canonical export**: export to the canonical site only when authorized; live article corrections belong on the site.
1. **Channel copy**: prepare channel adaptations in `posts/` (LinkedIn, X, Bluesky, Medium); channels are posted by hand, never automated or scheduled.

## Gotchas

- **Channel automation**: only canonical site export is automated; all other channels are manual copy-paste.
- **Immutable published copy**: once published, package copy is historical; live corrections stay on the published site.
- **Anti-slop enforcement**: avoid hype lead-ins, decorative lists, and empty corporate summaries; maintain the direct first-person voice.

## Documentation

- [Package layout](references/packages.md) · [Voice and identity](references/voice.md)
- Companion skills: [fmind-visuals](../fmind-visuals/SKILL.md) (diagram theme and brand), [d2](../d2/SKILL.md) (diagram source), [hugo](../hugo/SKILL.md) (docs sites), [release](../release/SKILL.md) (software releases), [typst](../typst/SKILL.md) (standalone documents).
