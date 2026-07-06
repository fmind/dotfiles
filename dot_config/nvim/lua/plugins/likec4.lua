-- Configure LikeC4 (Architecture-as-Code) plugin
return {
  {
    "likec4/likec4.nvim",
    event = { "BufReadPre", "BufNewFile" },
    opts = {},
  },
}
