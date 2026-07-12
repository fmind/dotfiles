-- General UI overrides and editor plugin fixes
return {
  -- Tokyo Night
  {
    "LazyVim/LazyVim",
    opts = {
      colorscheme = "tokyonight-moon",
    },
  },
  -- Which-key preset
  {
    "folke/which-key.nvim",
    opts = {
      preset = "classic",
    },
  },
  -- Fix refactoring.nvim error by ensuring async.nvim dependency is loaded
  {
    "ThePrimeagen/refactoring.nvim",
    dependencies = {
      "lewis6991/async.nvim",
    },
  },
}
