vim.g.mapleader = " "
vim.g.maplocalleader = " "

local map = vim.keymap.set

-- Motion: wrap-aware j/k (VSCode vim.normalModeKeyBindingsNonRecursive j→gj)
map("n", "j", "gj", { noremap = true })
map("n", "k", "gk", { noremap = true })
map("n", "B", "g^", { noremap = true })
map("n", "E", "g$", { noremap = true })
map("n", "Y", "y$", { noremap = true })

-- Visual: keep selection after indent (VSCode vim.visualModeKeyBindings)
map("x", "<", "<gv", { noremap = true })
map("x", ">", ">gv", { noremap = true })

-- Utility
map("n", "<CR>", ":", { noremap = true })
map("n", "U", "<C-r>", { noremap = true })
map("n", "gl", ":nohl<CR>", { noremap = true, desc = "Clear search highlight" })
map("c", "<C-p>", "<UP>", { noremap = true })
map("c", "<C-n>", "<DOWN>", { noremap = true })

-- Snacks: file/search (replaces Telescope bindings)
map("n", "<leader>e", function() Snacks.explorer() end, { desc = "File Explorer" })
map("n", "<leader>ff", function() Snacks.picker.files() end, { desc = "Find Files" })
map("n", "<C-p>", function() Snacks.picker.files() end, { desc = "Quick Open" })
map("n", "<leader>f", function() Snacks.picker.grep() end, { desc = "Live Grep" })
map("n", "<leader>b", function() Snacks.picker.buffers() end, { desc = "Buffers" })
map("n", "<leader>h", function() Snacks.picker.help() end, { desc = "Help Tags" })
map("n", "<leader>r", function() Snacks.picker.recent() end, { desc = "Recent Files" })
map("n", "<leader>/", function() Snacks.picker.grep_buffers() end, { desc = "Grep Buffers" })

-- Snacks: git
map("n", "<leader>gl", function() Snacks.lazygit() end, { desc = "LazyGit" })
map("n", "<leader>gf", function() Snacks.picker.git_files() end, { desc = "Git Files" })
map("n", "<leader>gc", function() Snacks.picker.git_log() end, { desc = "Git Log" })

-- Buffer navigation
map("n", "<leader>j", ":bnext<CR>", { noremap = true, desc = "Next Buffer" })
map("n", "<leader>k", ":bprevious<CR>", { noremap = true, desc = "Prev Buffer" })
map("n", "<leader>q", ":bdelete<CR>:bnext<CR>", { noremap = true, desc = "Delete Buffer" })

-- Terminal splits
map("n", "<leader>'", ":vsplit<CR>:terminal<CR>", { noremap = true, desc = "Terminal (vsplit)" })
map("n", '<leader>"', ":vsplit<CR>:terminal ipython<CR>", { noremap = true, desc = "IPython (vsplit)" })

-- Window management
map("n", "<A-o>", ":on<CR>", { noremap = true, desc = "Only Window" })
map("n", "<A-q>", ":close<CR>", { noremap = true, desc = "Close Window" })
map("n", "<A-s>", ":split<CR>", { noremap = true, desc = "Horizontal Split" })
map("n", "<A-v>", ":vsplit<CR>", { noremap = true, desc = "Vertical Split" })
map("n", "<A-t>", ":terminal<CR>", { noremap = true, desc = "Terminal" })
map("n", "<A-h>", "<C-w>h", { noremap = true })
map("n", "<A-j>", "<C-w>j", { noremap = true })
map("n", "<A-k>", "<C-w>k", { noremap = true })
map("n", "<A-l>", "<C-w>l", { noremap = true })

-- Terminal mode navigation
map("t", "<C-[>", "<C-\\><C-n>", { noremap = true })
map("t", "<A-h>", "<C-\\><C-N><C-w>h", { noremap = true })
map("t", "<A-j>", "<C-\\><C-N><C-w>j", { noremap = true })
map("t", "<A-k>", "<C-\\><C-N><C-w>k", { noremap = true })
map("t", "<A-l>", "<C-\\><C-N><C-w>l", { noremap = true })

-- Insert mode navigation
map("i", "<A-h>", "<C-\\><C-N><C-w>h", { noremap = true })
map("i", "<A-j>", "<C-\\><C-N><C-w>j", { noremap = true })
map("i", "<A-k>", "<C-\\><C-N><C-w>k", { noremap = true })
map("i", "<A-l>", "<C-\\><C-N><C-w>l", { noremap = true })

-- LSP (native vim.lsp, active after LspAttach)
vim.api.nvim_create_autocmd("LspAttach", {
  group = vim.api.nvim_create_augroup("lsp_keymaps", { clear = true }),
  callback = function(args)
    local opts = { buffer = args.buf, noremap = true }
    map("n", "gd", vim.lsp.buf.definition, vim.tbl_extend("force", opts, { desc = "Go to Definition" }))
    map("n", "gD", vim.lsp.buf.declaration, vim.tbl_extend("force", opts, { desc = "Go to Declaration" }))
    map("n", "gr", vim.lsp.buf.references, vim.tbl_extend("force", opts, { desc = "References" }))
    map("n", "gi", vim.lsp.buf.implementation, vim.tbl_extend("force", opts, { desc = "Implementation" }))
    map("n", "K", vim.lsp.buf.hover, vim.tbl_extend("force", opts, { desc = "Hover Docs" }))
    map("n", "<leader>ca", vim.lsp.buf.code_action, vim.tbl_extend("force", opts, { desc = "Code Action" }))
    map("n", "<leader>rn", vim.lsp.buf.rename, vim.tbl_extend("force", opts, { desc = "Rename Symbol" }))
    map("n", "<leader>ds", function() Snacks.picker.lsp_symbols() end, vim.tbl_extend("force", opts, { desc = "Document Symbols" }))
    map("n", "]d", vim.diagnostic.goto_next, vim.tbl_extend("force", opts, { desc = "Next Diagnostic" }))
    map("n", "[d", vim.diagnostic.goto_prev, vim.tbl_extend("force", opts, { desc = "Prev Diagnostic" }))
  end,
})

-- Terminal: no line numbers, start in insert mode
vim.api.nvim_create_autocmd("TermOpen", {
  group = vim.api.nvim_create_augroup("terminal_settings", { clear = true }),
  pattern = "*",
  callback = function()
    vim.cmd("startinsert")
    vim.opt_local.number = false
    vim.opt_local.relativenumber = false
  end,
})
