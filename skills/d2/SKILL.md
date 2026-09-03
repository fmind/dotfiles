---
name: d2
description: Create, theme, validate, and export D2 diagrams when an existing D2 source or a bespoke standalone composition beats the default Mermaid workflow.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/d2
  created: "2026-07-16"
  updated: "2026-09-03"
---

# D2 Diagram Standard

D2 is the specialist diagram tool: use it when the repository already owns `.d2` sources, the diagram needs bespoke composition (nested containers, boards, scenarios, layers, steps), a standalone SVG, PNG, PDF, PPTX, GIF, or ASCII export is the deliverable, or Mermaid cannot express the visual cleanly. Tool choice lives in [fmind-visuals](../fmind-visuals/SKILL.md).

## Workflow

1. **State the thesis**: the visual claim, target medium, and expected reading size.
1. **Write the smallest readable source**: prefer containers, explicit directions, and short labels over manual coordinates.
1. **Choose the layout**: start with `dagre`; try `elk` only when it produces a materially clearer topology (`d2 layout` lists the engines).
1. **Apply the palette**: an Fmind article diagram imports [diagram.d2](../fmind-visuals/references/diagram.d2) and uses its classes on a light surface; elsewhere start from a built-in theme (`d2 themes`) and apply the Fmind tokens through `theme-overrides` as in [fmind-theme](../fmind-visuals/references/fmind-theme.md).
1. **Format, validate, and render**, trusting the exit status rather than the presence of an output file:

   ```bash
   d2 fmt diagram.d2
   d2 validate diagram.d2
   d2 diagram.d2 diagram.svg
   ```

1. **Inspect the export**: clipping, crossings, font fallback, contrast, appendix behavior, and whether links or tooltips survive the target embedding.

## Gotchas

- **Raster by default**: prefer SVG for web pages and documentation; use PNG only for raster-only destinations.
- **Partial font sets**: supply all eight `--font-*` flags together with custom TTF files so missing weights do not silently fall back.
- **PNG and PDF need a browser**: D2 rasterizes through a bundled Playwright download; when it is unavailable, keep the SVG and rasterize it with [playwright](../playwright/SKILL.md) instead of pinning an obsolete browser.
- **Exported image as source**: keep `.d2` beside every generated artifact; the export is never the editable source.
- **Native rendering**: switch to [mermaid](../mermaid/SKILL.md) when the source must render inside GitHub or Slidev Markdown.

## Documentation

- [Language tour](https://d2lang.com/tour/intro/) · [CLI manual](https://d2lang.com/tour/man/) · [Themes and overrides](https://d2lang.com/tour/themes/)
- Companion skills: [fmind-visuals](../fmind-visuals/SKILL.md) (tool choice and brand), [mermaid](../mermaid/SKILL.md) (default diagrams), [playwright](../playwright/SKILL.md) (rasterizing SVG).
