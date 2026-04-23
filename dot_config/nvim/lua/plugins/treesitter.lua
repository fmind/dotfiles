-- Configure treesitter to add custom parsers and remove gitcommit
return {
	{
		"nvim-treesitter/nvim-treesitter",
		opts = function(_, opts)
			if type(opts.ensure_installed) == "table" then
				vim.list_extend(opts.ensure_installed, { "kdl", "just", "fish", "bash" })
				opts.ensure_installed = vim.tbl_filter(function(lang)
					return lang ~= "gitcommit"
				end, opts.ensure_installed)
			end
		end,
	},
}
