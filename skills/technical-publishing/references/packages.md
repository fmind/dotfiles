# Technical Publishing Packages

Technical publishing structures content into plain-file packages as the single source of truth.

## Package Layouts

### Articles

```text
articles/<slug>_<YYYY-MM-DD>/
├── draft.txt             # Author's raw notes; agents never edit it
├── article.md            # Working article markdown source
├── cover-prompt.md       # Image generation prompt for cover.png
├── diagrams/             # Explanatory visuals: D2 source beside render
│   ├── architecture.d2
│   └── architecture.png
├── images/               # Pictures (PNG)
│   └── cover.png         # Required cover image (opens the article)
└── posts/
    ├── seo.txt           # Metadata (title, description, tags, slug)
    ├── published.md      # Publication log (channel | date | URL)
    └── <channel files>   # linkedin.txt, x.txt, bluesky.txt, medium.md
```

The directory name defines the public slug and date: `<slug>` is the URL slug and `<YYYY-MM-DD>` is the public date. `posts/seo.txt` can override the slug.

Pipeline flow: `draft.txt -> article.md -> canonical site -> posts/`.

### Announcements

```text
announcements/<slug>_<YYYY-MM-DD>/
├── draft.txt             # News in author's words; agents never edit it
└── posts/                # Channel adaptations plus published.md
```

Pipeline flow: `draft.txt -> posts/`. Used for releases, awards, or talks with no standalone article.

### Media & Episodes

```text
podcasts/<id>_<name>/
├── draft.txt             # Initial concept and topic notes
├── sources.md            # Checked source set
├── prompts.md            # Production prompts
├── thumbnail-prompt.md   # Thumbnail generation brief
├── media/                # Inputs and generated media
└── posts/                # YouTube description, social posts, published.md
```
