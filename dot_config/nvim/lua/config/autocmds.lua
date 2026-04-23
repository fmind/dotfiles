vim.api.nvim_create_autocmd("FileType", {
	pattern = "markdown",
	callback = function()
		vim.opt_local.textwidth = 0 -- disable hard wrap
		vim.opt_local.formatoptions:remove("t") -- disable auto-wrap
	end,
})
