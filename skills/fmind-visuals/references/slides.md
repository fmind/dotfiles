---
theme: default
title: Presentation title
author: Médéric Hurier
colorSchema: dark
aspectRatio: 16/9
canvasWidth: 980
selectable: true
fonts:
  sans: Inter
  serif: Outfit
  mono: JetBrains Mono
  provider: none
themeConfig:
  primary: "#646CFF"
defaults:
  transition: slide-left
---

<img src="/brand/logo.webp" alt="Fmind" class="fmind-logo" />

# Presentation title

## One concrete tension, decision, or mechanism

<div class="fmind-meta">
  Médéric Hurier · www.fmind.dev
</div>

<!--
State the audience, why this matters now, and the single takeaway.
-->

---
layout: default
---

# One idea per slide

- Start from concrete friction.
- Show the mechanism or evidence.
- End with a decision boundary.

---
layout: default
---

# Portable diagrams use Mermaid

```mermaid
---
config:
  fontFamily: "ui-sans-serif, system-ui, sans-serif"
  flowchart:
    diagramPadding: 16
  theme: base
  themeVariables:
    darkMode: true
    background: "#0F172A"
    primaryColor: "#1E293B"
    primaryTextColor: "#F8FAFC"
    primaryBorderColor: "#646CFF"
    lineColor: "#CBD5E1"
    textColor: "#F8FAFC"
---
flowchart LR
  Friction --> Mechanism --> Decision
```

---
layout: end
---

# The useful next move

www.fmind.dev
