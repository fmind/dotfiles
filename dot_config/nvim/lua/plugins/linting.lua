return {
  {
    "mfussenegger/nvim-lint",
    opts = function(_, opts)
      opts.linters_by_ft = opts.linters_by_ft or {}
      opts.linters_by_ft.markdown = {}

      -- nvim-lint's default golangcilint args already emit JSON to stdout with
      -- --path-mode=abs and a trailing path argument. Override only that final
      -- path element to lint the package directory containing go.mod instead of
      -- the single file, avoiding false-positive typecheck errors (undefined symbols).
      local ok, lint = pcall(require, "lint")
      if ok and lint.linters and lint.linters.golangcilint then
        local args = lint.linters.golangcilint.args
        args[#args] = function()
          local bufname = vim.api.nvim_buf_get_name(0)
          local dirname = vim.fs.dirname(bufname)
          local go_mod = vim.fs.find("go.mod", { path = dirname, upward = true })[1]
          return go_mod and dirname or bufname
        end
      end

      return opts
    end,
  },
}
