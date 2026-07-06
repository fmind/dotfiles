-- Configure Language Servers (LSP)
return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      servers = {
        templ = {},
        tailwindcss = {
          filetypes_include = { "templ" },
          settings = {
            tailwindCSS = {
              includeLanguages = {
                templ = "html",
              },
            },
          },
        },
      },
    },
  },
}
