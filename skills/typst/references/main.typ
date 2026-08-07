// Document metadata and base style: tune once per document.
#set document(title: "Document Title", author: "Médéric Hurier (Fmind)")
#set page(paper: "a4", margin: (x: 2.4cm, y: 2.6cm), numbering: "1 / 1")
#set text(font: "Libertinus Serif", size: 11pt, lang: "en")
#set par(justify: true)
#set heading(numbering: "1.1")

#align(center)[
  #text(size: 20pt, weight: "bold")[Document Title]

  #v(0.4em)
  Médéric Hurier (Fmind) · #datetime.today().display("[month repr:long] [day], [year]")
]

#outline(depth: 2)

= Introduction

Start writing here. Typst compiles in milliseconds — keep `mise run watch`
running for a live PDF preview while editing.

== Math

Inline math like $e = m c^2$ and display math:

$ integral_0^oo e^(-x^2) dif x = sqrt(pi) / 2 $

== Citations

// Add refs.bib (BibTeX) or refs.yaml (Hayagriva), cite with @key, then close
// the document with: #bibliography("refs.bib")
