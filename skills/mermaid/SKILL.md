---
name: mermaid
description: Create, theme, validate, and embed Mermaid diagrams, the default diagram format for GitHub, Hugo, Slidev, and docs. Use for any new diagram unless D2 or LikeC4 applies.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/mermaid
  created: "2026-07-16"
  updated: "2026-09-03"
---

# Mermaid Diagram Standard

Mermaid is the default diagram format because the same editable text renders in GitHub Markdown, Hugo sites (Hextra bundles it), and Slidev decks. Keep the source portable, reviewable, and close to the prose it explains; the choice between Mermaid, [D2](../d2/SKILL.md), LikeC4, or no diagram lives in [fmind-visuals](../fmind-visuals/SKILL.md).

## Workflow

1. **State the thesis**: one visual claim and the reader decision it supports; omit the diagram when prose, code, a list, or a table is more direct.
1. **Pick a stable type**: prefer `flowchart`, `sequenceDiagram`, `stateDiagram-v2`, `classDiagram`, and `erDiagram`; avoid a newly released type until every target renderer's Mermaid version supports it (GitHub reports its version from a block containing `info`).
1. **Write portable source**: a fenced `mermaid` block when the diagram belongs to one Markdown document, a `.mmd` file when it is reused or rendered independently; keep labels short, direction intentional, and the node count readable without zooming.
1. **Configure in frontmatter**: put configuration in Mermaid frontmatter, never in `%%{init: ...}%%` directives or renderer-specific fence options; apply the Fmind theme from [fmind-theme](../fmind-visuals/references/fmind-theme.md) when the work represents Médéric or `www.fmind.dev`.
1. **Validate and render**:

   ```bash
   mmdc -i diagram.mmd -o diagram.svg -b transparent
   ```

1. **Inspect the render**: clipping, line crossings, contrast, spelling, label density, and mobile or slide readability.
1. **Ship the source**: keep the Mermaid text; add SVG only when the target cannot render Mermaid and PNG only when a platform requires raster output, and add `accTitle` and `accDescr` or surrounding prose so the conclusion survives without the image.

## Gotchas

- **No browser found**: `mmdc` fails with `Could not find chrome-headless-shell` when the Puppeteer download is absent; point it at system Chrome: `PUPPETEER_EXECUTABLE_PATH=/usr/bin/google-chrome mmdc -i diagram.mmd -o diagram.svg`.
- **Clipped labels**: set one renderer-stable font stack through root-level `config.fontFamily`; late-loading web fonts change label measurements after layout and clip inside fixed bounds.
- **Deprecated option**: `flowchart.htmlLabels` is deprecated; do not add it to new diagrams.
- **Non-portable features**: keep remote images, custom JavaScript, click callbacks, and renderer plugins out of shared source.
- **Dense diagrams**: split into views instead of shrinking labels; left-to-right for slide-sized processes, top-to-bottom for document hierarchies.
- **Invented structure**: preserve source terminology; never add metrics, components, trust boundaries, or causal links the evidence does not show.

## Documentation

- [Syntax reference](https://mermaid.js.org/intro/syntax-reference.html) · [Theming](https://mermaid.js.org/config/theming) · [Mermaid CLI](https://github.com/mermaid-js/mermaid-cli)
- Companion skills: [fmind-visuals](../fmind-visuals/SKILL.md) (tool choice and Fmind theme), [d2](../d2/SKILL.md) (bespoke compositions), [hugo](../hugo/SKILL.md) (Hextra sites).
