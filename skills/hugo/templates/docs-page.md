---
title: Getting Started
weight: 1
math: true # opt in per page so KaTeX assets load only where needed
---

## Install

```bash
mise run install
```

{{< callout type="info" >}} Hugo modules pin the theme in `go.mod`; run `hugo mod get -u` to upgrade. {{< /callout >}}

## Diagrams

```mermaid
graph LR
  A[Markdown] --> B[Hugo] --> C[Static Site]
```

## Math

Inline math renders server-side: \(E = mc^2\).

$$ \int_0^1 x^2 \, dx = \frac{1}{3} $$
