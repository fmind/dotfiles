vim.g.loaded_perl_provider = 0
vim.g.loaded_ruby_provider = 0

local node_host = vim.fn.exepath("neovim-node-host")
if node_host ~= "" then
	vim.g.node_host_prog = node_host
end

local pynvim_host = vim.fn.exepath("pynvim-python")
if pynvim_host ~= "" then
	local python_host = vim.fs.joinpath(vim.fn.fnamemodify(pynvim_host, ":h"), "python")
	if vim.fn.executable(python_host) == 1 then
		vim.g.python3_host_prog = python_host
	end
end
