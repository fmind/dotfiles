return {
	{
		"stevearc/conform.nvim",
		opts = {
			formatters = {
				prettier = {
					prepend_args = function(_, ctx)
						if vim.bo[ctx.buf].filetype == "markdown" then
							return { "--prose-wrap", "never" }
						end
						return {}
					end,
				},
			},
		},
	},
}
