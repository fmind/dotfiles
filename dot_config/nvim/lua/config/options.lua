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
  -- xclip exits non-zero with "Error: target STRING not available" whenever a selection
  -- has no owner or holds non-text data, which Neovim surfaces as a clipboard error on
  -- every paste. Read through a Lua function so an unavailable selection is simply an
  -- empty register instead of an error message.
  local function paste(selection)
    return function()
      local lines = vim.fn.systemlist({ "xclip", "-selection", selection, "-o" })
      return vim.v.shell_error == 0 and lines or {}
    end
  end

  vim.g.clipboard = {
    name = "xclip",
    copy = {
      ["+"] = "xclip -selection clipboard",
      ["*"] = "xclip -selection primary",
    },
    paste = {
      ["+"] = paste("clipboard"),
      ["*"] = paste("primary"),
    },
    cache_enabled = 1,
  }
end
