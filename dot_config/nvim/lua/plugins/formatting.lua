return {
  {
    "stevearc/conform.nvim",
    opts = function(_, opts)
      opts.formatters_by_ft = opts.formatters_by_ft or {}
      opts.formatters_by_ft.yaml = { "dprint" }
      opts.formatters_by_ft.templ = { "templ" }

      return opts
    end,
  },
}
