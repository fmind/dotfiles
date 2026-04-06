local opt = vim.opt

opt.shell = "bash"

-- Core behaviors
opt.undofile = true           -- retain undo history (saved in ~/.local/state/nvim/undo/)
opt.updatetime = 250          -- faster swap file and git signs updates
opt.hidden = true
opt.confirm = true
opt.autoread = true
opt.autowrite = true

opt.splitbelow = true
opt.splitright = true

opt.autoindent = true
opt.formatoptions:remove("cro")

opt.foldmethod = "indent"
opt.foldlevelstart = 99

opt.tabstop = 4
opt.expandtab = true
opt.shiftround = true
opt.shiftwidth = 4
opt.softtabstop = 4

opt.number = true
opt.relativenumber = true
opt.cursorline = true         -- highlight current line
opt.signcolumn = "yes"        -- always show signcolumn

opt.termguicolors = true      -- enable 24-bit RGB colors

opt.wildmode = "list:longest,full"
opt.completeopt = "menuone,noselect"

opt.ignorecase = true
opt.smartcase = true
opt.incsearch = true
opt.hlsearch = true

opt.clipboard = "unnamedplus"

opt.linebreak = true
opt.shortmess:append("I")
opt.scrolloff = 15            -- synced from VSCode cursorSurroundingLines: 15

opt.smoothscroll = true       -- smooth scrolling (Neovim 0.10+, VSCode smoothScrolling)
opt.fixendofline = true       -- always end file with newline (VSCode insertFinalNewline)

-- Show trailing whitespace (VSCode renderWhitespace: "trailing")
opt.list = true
opt.listchars = { trail = "·", tab = "→ ", nbsp = "␣" }

-- Highlighted yank (VSCode vim.highlightedyank)
vim.api.nvim_create_autocmd("TextYankPost", {
  group = vim.api.nvim_create_augroup("highlight_yank", { clear = true }),
  pattern = "*",
  callback = function()
    vim.highlight.on_yank({ higroup = "IncSearch", timeout = 200 })
  end,
})

-- Strip trailing whitespace on save (VSCode trimTrailingWhitespace)
vim.api.nvim_create_autocmd("BufWritePre", {
  group = vim.api.nvim_create_augroup("trim_whitespace", { clear = true }),
  pattern = "*",
  callback = function()
    if not vim.bo.modifiable then return end
    local pos = vim.api.nvim_win_get_cursor(0)
    vim.cmd([[%s/\s\+$//e]])
    vim.api.nvim_win_set_cursor(0, pos)
  end,
})
