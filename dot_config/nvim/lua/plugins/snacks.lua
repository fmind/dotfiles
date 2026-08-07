return {
  {
    "folke/snacks.nvim",
    opts = {
      scroll = { enabled = false },
      image = { enabled = false },
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
