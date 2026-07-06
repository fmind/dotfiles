-- Markdown: disable hard wrapping. Soft wrap/linebreak are already set globally
-- in options.lua, so only the markdown-specific bits live here.
vim.api.nvim_create_autocmd("FileType", {
	pattern = "markdown",
	callback = function()
		vim.opt_local.formatoptions:remove("t") -- disable auto-wrap
		vim.opt_local.textwidth = 0 -- disable hard wrap
	end,
})
