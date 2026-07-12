---
name: html-slides
description: Build a presentation as one self-contained, zero-dependency HTML file (16:9, keyboard + touch nav), authored directly by the agent. Use when creating slides, a deck, or a talk from notes or a document.
license: MIT
metadata:
  author: Médéric HURIER (Fmind)
  source: github.com/fmind/dotfiles/tree/main/skills/html-slides
  created: 2026-07-07
  updated: 2026-07-09
---

# HTML Slides Standard

Produce a presentation as a **single, self-contained `.html` file** — inline CSS/JS, no npm, no build, no CDN. It opens in any browser, works offline, and deep-links per slide. A coding agent writes HTML fluently, so it authors the deck directly — full design control, zero dependencies. _Dependencies are debt; a single HTML file still works in 10 years._

## 1. Workflow

1. **Gather content**: collect the raw material (notes, README, doc, transcript) and the goal — audience, length, one takeaway per slide. Outline the slide list before writing markup.
1. **Set the theme**: copy [template.html](references/template.html) to `deck.html`, set `<title>`, and edit the `:root` palette. Change it as a set — `--bg` (letterbox) **and** `--slide-bg` (slide) together, plus `--fg`, `--muted`, `--accent`, `--code-bg`, `--font`; for a dark theme set `--bg`/`--slide-bg` dark or the letterbox stays white. Pick a distinctive palette — avoid the generic look.
1. **Write slides**: one `<section class="slide">` per idea inside the fixed 1280×720 canvas. Use the built-in `.center`, `.cols`, `.kicker`, `.fragment` helpers. Keep ~one message per slide; let whitespace breathe.

## 2. Design on a Fixed Canvas

Every slide is a **1280×720 px** canvas; the stage scales to fit the window (letterboxed) via JS, so layout is pixel-predictable. Design to that fixed size — do not use viewport units for layout. Slide-authoring conventions:

- **Type scale** is preset (`h1` 64 / `h2` 44 / body 28). Trust it; resist shrinking text to cram content — split the slide instead.
- **Progressive reveal**: add `class="fragment"` to any element to reveal it step-by-step on →/Space before the deck advances.
- **Code**: use `<pre><code>` (monospace token preset). Paste real, minimal snippets — no live highlighting runtime by design.
- **Images**: keep the deck self-contained — embed as `data:` URIs: `<img src="data:image/png;base64,…">` (base64-encode the file). For large or many images, weigh the file-size cost against portability.
- **Per-slide background**: override the theme on one slide with `<section class="slide" style="background:#0b1021">` (or a per-slide class).
- **Offline fonts**: the default is a system-font stack (zero network). Only add a web font if the deck will always have connectivity, and note the trade-off.

## 3. Controls

- **Keyboard**: →/Space/PageDown next · ←/PageUp prev · Home/End first/last · `F` fullscreen.
- **Touch**: swipe left/right on touch devices for next/prev.
- **Print**: Ctrl+P prints each slide as one exact 16:9 page (all fragments auto-revealed).
- **Deep-link**: jump to a slide with `deck.html#4`.

## 4. Gotchas

- **Keep it one file**: inline everything. No CDN links, no external assets. Portability is the whole point.
- **Contrast & size for a room**: assume a projector — high contrast, nothing smaller than the 28px body preset.

## Documentation

- [template.html](references/template.html) — the canonical single-file deck (theme tokens, layouts, fragments, touch nav).
