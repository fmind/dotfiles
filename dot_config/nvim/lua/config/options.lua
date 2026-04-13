vim.g.loaded_perl_provider = 0
vim.g.loaded_ruby_provider = 0

local node_host = vim.fn.exepath("neovim-node-host")
if node_host ~= "" then
	vim.g.node_host_prog = node_host
end

-- Markdown: soft wrap, no hard wrap
vim.api.nvim_create_autocmd("FileType", {
	pattern = "markdown",
	callback = function()
		vim.opt_local.wrap = true -- enable soft wrap
		vim.opt_local.linebreak = true -- break lines at words
		vim.opt_local.textwidth = 0 -- disable hard wrap
		vim.opt_local.formatoptions:remove("t") -- disable auto-wrap
	end,
})
