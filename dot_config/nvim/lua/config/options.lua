vim.g.loaded_perl_provider = 0
vim.g.loaded_ruby_provider = 0

local node_host = vim.fn.exepath("neovim-node-host")
if node_host ~= "" then
	vim.g.node_host_prog = node_host
end
