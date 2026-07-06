local map = vim.keymap.set

-- Access literal (physical) line movement via gj/gk.
-- Plain j/k are left to LazyVim's count-aware default (v:count == 0 ? 'gj' : 'j'),
-- so {count}j/{count}k keep jumping real lines with relativenumber.
map({ "n", "v" }, "gj", "j", { silent = true })
map({ "n", "v" }, "gk", "k", { silent = true })

-- Close buffer and switch to the next one
local function bdelete_next()
  local buf = vim.api.nvim_get_current_buf()
  vim.cmd("bnext")
  if vim.api.nvim_get_current_buf() == buf then
    vim.cmd("new")
  end
  if vim.api.nvim_buf_is_valid(buf) then
    pcall(vim.api.nvim_buf_delete, buf, {})
  end
end
map("n", "<leader>bd", bdelete_next, { desc = "Delete Buffer (next)" })
