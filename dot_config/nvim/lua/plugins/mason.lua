return {
  {
    "mason-org/mason.nvim",
    opts = function(_, opts)
      if type(opts.ensure_installed) == "table" then
        opts.ensure_installed = vim.tbl_filter(function(tool)
          -- Mason package names can differ from their binary (e.g. delve → dlv);
          -- map known mismatches so mise-provided tools are correctly skipped.
          local bin = ({ delve = "dlv" })[tool] or tool
          return vim.fn.executable(bin) == 0
        end, opts.ensure_installed)
      end
      return opts
    end,
  },
}
