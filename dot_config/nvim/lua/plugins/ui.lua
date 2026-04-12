return {
  -- Remove lualine symbols and powerline separators
  {
    "nvim-lualine/lualine.nvim",
    opts = function(_, opts)
      opts.options.icons_enabled = false
      opts.options.component_separators = { left = "|", right = "|" }
      opts.options.section_separators = { left = "", right = "" }
    end,
  },

  -- Disable icons for bufferline
  {
    "akinsho/bufferline.nvim",
    opts = function(_, opts)
      opts.options.show_buffer_icons = false
      opts.options.show_buffer_close_icons = false
      opts.options.show_close_icon = false
      opts.options.show_tab_indicators = true
    end,
  },

  -- Clean up LazyVim default icon sets
  {
    "LazyVim/LazyVim",
    opts = {
      icons = {
        diagnostics = {
          Error = "E:",
          Warn  = "W:",
          Hint  = "H:",
          Info  = "I:",
        },
        git = {
          added    = "+",
          modified = "~",
          removed  = "-",
        },
        kinds = {
          Array         = "",
          Boolean       = "",
          Class         = "",
          Color         = "",
          Constant      = "",
          Constructor   = "",
          Copilot       = "",
          Enum          = "",
          EnumMember    = "",
          Event         = "",
          Field         = "",
          File          = "",
          Folder        = "",
          Function      = "",
          Interface     = "",
          Key           = "",
          Keyword       = "",
          Method        = "",
          Module        = "",
          Namespace     = "",
          Null          = "",
          Number        = "",
          Object        = "",
          Operator      = "",
          Package       = "",
          Property      = "",
          Reference     = "",
          Snippet       = "",
          String        = "",
          Struct        = "",
          Text          = "",
          TypeParameter = "",
          Unit          = "",
          Value         = "",
          Variable      = "",
        },
      },
    },
  },

  -- Override snacks explorer explicitly to not use icons
  {
    "folke/snacks.nvim",
    opts = {
      explorer = {
        replace_netrw = true,
      },
      indent = {
        char = "|",
      },
    },
  },

  -- Trouble symbol remover
  {
    "folke/trouble.nvim",
    opts = {
      icons = {
        indent = {
          top         = "| ",
          middle      = "|-",
          last        = "`-",
          fold_open   = "v ",
          fold_closed = "> ",
          ws          = "  ",
        },
        folder_closed = "> ",
        folder_open   = "v ",
        kinds = {},
      },
    },
  },

  -- Down-grade mini.icons explicitly
  {
    "nvim-mini/mini.icons",
    opts = {
      style = "ascii",
    },
  },
}
