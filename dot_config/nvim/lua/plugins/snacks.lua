return {
  {
    "folke/snacks.nvim",
    opts = {
      scroll = { enabled = false },
      -- Inline images in the buffer. Ghostty speaks the kitty graphics
      -- protocol, and PNG needs no `magick`, so article diagrams render where
      -- their ![...](diagrams/*.png) reference sits.
      image = { enabled = true },
      picker = {
        sources = {
          files = {
            hidden = true,
            ignored = false,
            exclude = {
              "*.png",
              "*.webp",
              "*.jpg",
              "*.jpeg",
              "*.gif",
              "*.ico",
              "*.bmp",
              "*.tiff",
              "*.avif",
              "*.heic",
              "*.pdf",
              "*.zip",
              "*.tar",
              "*.gz",
              "*.7z",
              "*.mp3",
              "*.mp4",
              "*.mov",
              "*.webm",
              "*.woff",
              "*.woff2",
              "*.ttf",
              "*.eot",
              "*.otf",
            },
          },
        },
      },
    },
  },
}
