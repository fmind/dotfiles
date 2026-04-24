local map = vim.keymap.set

-- Movement on visual lines
map({ "n", "v" }, "j", "gj", { silent = true })
map({ "n", "v" }, "k", "gk", { silent = true })
map({ "n", "v" }, "gj", "j", { silent = true })
map({ "n", "v" }, "gk", "k", { silent = true })

-- Quick "clear search"
map("n", "gl", "<cmd>nohlsearch<cr>", { silent = true, desc = "Clear search highlight" })
