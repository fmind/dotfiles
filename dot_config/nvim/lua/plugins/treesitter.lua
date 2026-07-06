-- Configure treesitter to add custom parsers and remove gitcommit on older hosts
-- gitcommit is filtered out because building it requires a newer GLIBC (>= 2.39)
-- than some legacy systems provide.
return {
  {
    "nvim-treesitter/nvim-treesitter",
    opts = function(_, opts)
      if type(opts.ensure_installed) == "table" then
        vim.list_extend(opts.ensure_installed, { "kdl", "just", "fish", "templ" })

        -- Dynamically filter out gitcommit only if host's GLIBC is older than 2.39
        local handle = io.popen("ldd --version 2>/dev/null | head -n 1")
        local result = handle and handle:read("*a") or ""
        if handle then handle:close() end
        local version_str = string.match(result, "GLIBC%s+([%d%.]+)")
        if version_str then
          local major, minor = string.match(version_str, "^(%d+)%.(%d+)")
          major, minor = tonumber(major) or 0, tonumber(minor) or 0
          if major < 2 or (major == 2 and minor < 39) then
            opts.ensure_installed = vim.tbl_filter(function(lang)
              return lang ~= "gitcommit"
            end, opts.ensure_installed)
          end
        end
      end
      return opts
    end,
  },
}
