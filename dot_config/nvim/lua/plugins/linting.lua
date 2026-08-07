return {
  {
    "mfussenegger/nvim-lint",
    opts = function(_, opts)
      opts.linters_by_ft = opts.linters_by_ft or {}
      opts.linters_by_ft.markdown = {}

      local ok, lint = pcall(require, "lint")
      if ok and lint.linters and lint.linters.golangcilint then
        -- golangci-lint only understands Go packages. Agent skill references (and any
        -- other snippet outside a module) have no go.mod above them, so it fails with
        -- exit code 3 on a typecheck error instead of reporting real findings. Skip
        -- those buffers entirely rather than surfacing the failure to the user.
        lint.linters.golangcilint.condition = function(ctx)
          return vim.fs.root(ctx.filename, "go.mod") ~= nil
        end

        -- nvim-lint's default golangcilint args already emit JSON to stdout with
        -- --path-mode=abs and a trailing path argument. Override only that final
        -- path element to lint the package directory instead of the single file,
        -- avoiding false-positive typecheck errors (undefined symbols).
        local args = lint.linters.golangcilint.args
        args[#args] = function()
          return vim.fs.dirname(vim.api.nvim_buf_get_name(0))
        end
      end

      return opts
    end,
  },
}
