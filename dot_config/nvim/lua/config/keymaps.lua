local map = vim.keymap.set

-- Reverse visual line movement
-- LazyVim already maps j/k to gj/gk smartly
map({ "n", "v" }, "gj", "j", { silent = true })
map({ "n", "v" }, "gk", "k", { silent = true })
