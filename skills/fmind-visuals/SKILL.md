---
name: fmind-visuals
description: Apply the Fmind visual identity and route slide or diagram work to Slidev, Mermaid, LikeC4, or D2. Use for Fmind talks, decks, article diagrams, site assets.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dot/tree/main/skills/fmind-visuals
  created: "2026-07-16"
  updated: "2026-09-03"
---

# Fmind Visual Communication

Create calm, exact, spacious technical visuals that feel native to `www.fmind.dev`. The bundled [Fmind theme reference](references/fmind-theme.md) and [diagram template](references/diagram.d2) define the brand tokens, palette, and diagram classes; article production itself belongs to [technical-publishing](../technical-publishing/SKILL.md).

## Canonical Tool Choice

| Need                                                        | Tool                           | Boundary                                                  |
| ----------------------------------------------------------- | ------------------------------ | --------------------------------------------------------- |
| Slides, talks, workshops, LinkedIn documents                | [Slidev](https://sli.dev)      | Default for every new deck                                |
| Flow, sequence, state, class, ER, compact technical diagram | [Mermaid](../mermaid/SKILL.md) | Default for every new diagram                             |
| Fmind article diagram                                       | [D2](../d2/SKILL.md)           | Import [diagram.d2](references/diagram.d2), light surface |
| Durable architecture model with multiple generated views    | [LikeC4](https://likec4.dev/)  | Use when the model, not one image, is the source of truth |
| Existing D2 source or bespoke standalone composition        | [D2](../d2/SKILL.md)           | Specialist fallback                                       |

Do not create a custom HTML deck, Typst deck, PowerPoint source, or generated raster diagram unless the user explicitly requests that format or an existing project requires it.

## Brand Contract

The dark tokens below apply to decks and site assets; an Fmind article diagram draws on a light surface with the classes of [diagram.d2](references/diagram.d2) instead, because the site toggles between light and dark themes.

- Heading font: Outfit Variable.
- Body font: Inter Variable.
- Background: `#0F172A`.
- Panel: `#1E293B`.
- Foreground: `#F8FAFC`.
- Muted: `#CBD5E1`.
- Primary accent: `#646CFF`.
- Border: `#334155`.
- Voice: calm, exact, pragmatic, technically grounded, and explicit about trade-offs.
- Use canonical font files (Outfit Variable, Inter Variable); copy them into the deliverable rather than linking to an external path.
- Use the Bleeding Agent palette only when the user explicitly asks for that sub-brand.

## Workflow

### Decks

1. **Scaffold with pnpm**: keep Slidev, Vue, the default theme, and `playwright-chromium` project-local; start from [package.json.template](references/package.json.template), [pnpm-workspace.yaml](references/pnpm-workspace.yaml), [slides.md](references/slides.md), and [style.css](references/style.css), then copy the logo and WOFF2 fonts into `public/brand/`.
1. **Keep the DOMPurify override**: until Monaco no longer pins a vulnerable release; verify any removal with `pnpm audit`.
1. **One idea per slide**: one claim, mechanism, decision, or artifact; split dense content instead of shrinking type.
1. **Embed diagrams**: Mermaid directly for ordinary diagrams; exported LikeC4 or D2 SVGs only when their specialist boundary applies.
1. **Run, build, export**:

   ```bash
   slidev slides.md
   slidev build slides.md
   slidev export slides.md
   ```

1. **Inspect every view**: browser, projector-sized, and exported; prefer Slidev's browser exporter for review PNGs or PPTX and keep CLI PDF export for automation.

### Diagrams

1. **Start with Mermaid**: apply the portable Fmind frontmatter from [fmind-theme](references/fmind-theme.md); use [d2](../d2/SKILL.md) or LikeC4 only when their composition or model advantages outweigh the loss of direct Markdown rendering.
1. **Keep source beside exports**: near the prose or deck that owns the claim; export SVG only for destinations that cannot render Mermaid.

## Gotchas

- **Interactive success is not export success**: inspect the PDF or PNG; fixed bounds clip late-loading fonts.
- **Decoration**: every node and slide carries one evidence-backed thesis; remove decorative nodes, gradients, and generic AI imagery.
- **Accessibility**: diagrams need a prose equivalent or alt text; text stays legible on a laptop, projector, mobile preview, and exported page.
- **Private paths**: copy brand assets into the deliverable; never link local workspace paths from a published artifact.

## Official Skills

Upstream: `slidevjs/slidev` (decks) and `likec4/likec4` (architecture models). List the current release, then install what the task needs at project scope after reviewing the snapshot (see [agent-skills](../agent-skills/SKILL.md)):

```bash
skills add slidevjs/slidev --list
skills add slidevjs/slidev --skill <name> -y
skills add likec4/likec4 --list
skills add likec4/likec4 --skill <name> -y
```

## Documentation

- [Fmind website](https://www.fmind.dev/) · [Slidev](https://sli.dev) · [Mermaid](https://mermaid.js.org/) · [LikeC4](https://likec4.dev/) · [D2](https://d2lang.com/)
- Companion skills: [mermaid](../mermaid/SKILL.md) (default diagrams), [d2](../d2/SKILL.md) (specialist diagrams), [technical-publishing](../technical-publishing/SKILL.md) (Fmind articles), [agent-skills](../agent-skills/SKILL.md) (upstream skill install).
