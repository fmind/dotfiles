---
name: technical-publishing
description: "Publish source-grounded technical articles across draft, canonical site, export, and public verification. Not software releases or standalone documents."
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/technical-publishing
  created: 2026-08-30
  updated: 2026-08-30
---

# Publish Technical Articles

Carry one article from an evidence-backed source draft to its canonical public form without losing authorship, links, metadata, or proof boundaries. Use [hugo](../hugo/SKILL.md) for site mechanics and [typst](../typst/SKILL.md) for standalone reports.

## Workflow

1. **Resolve the source of truth:** Identify the canonical repository, article identifier, target audience, publication surfaces, and existing public URL before editing. Preserve unrelated drafts and generated artifacts.
1. **Define the claim:** Write the reader's problem, the article's useful conclusion, and the evidence needed to support it. Remove claims that cannot be verified or clearly label them as opinion or inference.
1. **Research current sources:** Prefer installed source, standards, primary documentation, and reproducible experiments. Record dates and versions for facts likely to drift; paraphrase within source and copyright limits.
1. **Draft for one reading path:** Lead with the outcome, keep code minimal and runnable, explain trade-offs at the decision point, and provide safe failure and recovery guidance.
1. **Validate the artifact:** Run formatting, links, code examples, site build, metadata, images, accessibility, canonical URL, and the repository's full gate. Inspect the rendered page, not only Markdown.
1. **Prepare each surface:** Derive exports from the canonical source, adapt only platform-specific metadata and length, and verify that links still point back to the canonical article.
1. **Publish only with authority:** Treat commits, pushes, deploys, cross-posts, newsletters, and social messages as distinct external mutations. Never infer them from a request to edit or review.
1. **Verify publicly:** After publication, fetch the public page and confirm title, body, code, images, canonical metadata, links, and deployment revision. Keep local, deployed, indexed, and analytics evidence separate.

## Handoff

Report the canonical source and URL, supported claims, validation performed, exact surfaces changed, and any manual publication, indexing, or analytics check that remains.
