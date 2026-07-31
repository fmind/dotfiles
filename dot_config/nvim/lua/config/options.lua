local opt = vim.opt

-- Enable soft wrapping
opt.wrap = true
opt.linebreak = true

-- Substitute options
opt.gdefault = true

-- Keymap timeouts
opt.timeoutlen = 400

-- System clipboard synchronization
opt.clipboard:append("unnamedplus")

-- Prefer xclip on Linux/ChromeOS to avoid wl-clipboard hanging in Wayland containers
if vim.fn.executable("xclip") == 1 then
  vim.g.clipboard = {
    name = "xclip",
    copy = {
      ["+"] = "xclip -selection clipboard",
      ["*"] = "xclip -selection primary",
    },
    paste = {
      ["+"] = "xclip -selection clipboard -o",
      ["*"] = "xclip -selection primary -o",
    },
    cache_enabled = 1,
  }
end
