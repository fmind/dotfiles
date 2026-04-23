return {
  {
    "nvim-mini/mini.icons",
    opts = {
      style = "ascii",
    },
  },
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        icons_enabled = false,
      },
    },
  },
  {
    "nvim-neo-tree/neo-tree.nvim",
    opts = {
      default_component_configs = {
        icon = {
          folder_empty = "[-]",
          folder_empty_open = "[-]",
          folder_closed = "[+]",
          folder_open = "[-]",
          default = "[ ]",
        },
        name = {
          use_git_status_colors = true,
        },
      },
    },
  },
}
