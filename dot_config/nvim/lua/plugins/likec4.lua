return {
  {
    "likec4/likec4.nvim",
    -- The plugin ships no `lua/` module (only ftdetect/ftplugin/syntax/lsp),
    -- so `opts`/`config = true` makes lazy.nvim fail on `require("likec4")`.
    -- Load it on the filetype instead: lazy.nvim sources the plugin's
    -- ftdetect at startup when `ft` is set, so `.c4` files still trigger it.
    ft = "likec4",
    -- The `likec4` LSP binary comes from mise (npm:@likec4/lsp), not `build`.
  },
}
