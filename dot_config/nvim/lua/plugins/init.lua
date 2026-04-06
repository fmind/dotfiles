return {
  -- ─── Theme ───────────────────────────────────────────────────────────────
  {
    "catppuccin/nvim",
    name = "catppuccin",
    lazy = false,
    priority = 1000,
    opts = {
      flavour = "latte",
      transparent_background = false,
      show_end_of_buffer = false,
      term_colors = true,
      dim_inactive = {
        enabled = false,
      },
      styles = {
        comments = { "italic" },
        conditionals = { "italic" },
        loops = {},
        functions = {},
        keywords = { "italic" },
        strings = {},
        variables = {},
        numbers = {},
        booleans = {},
        properties = {},
        types = {},
        operators = {},
      },
      integrations = {
        blink_cmp = true,
        gitsigns = true,
        flash = true,
        noice = true,
        notify = true,
        which_key = true,
        snacks = true,
        treesitter = true,
        render_markdown = true,
        trouble = true,
      },
    },
    config = function(_, opts)
      require("catppuccin").setup(opts)
      vim.cmd.colorscheme("catppuccin")
    end,
  },

  -- ─── Status Line ─────────────────────────────────────────────────────────
  {
    "nvim-lualine/lualine.nvim",
    opts = {
      options = {
        theme = "catppuccin",
        icons_enabled = false,
      },
    },
  },

  -- ─── Snacks (picker + dashboard + notifications + lazygit + more) ───────
  {
    "folke/snacks.nvim",
    priority = 1000,
    lazy = false,
    opts = {
      picker = {
        enabled = true,
        icons = { enabled = false },
      },
      explorer = {
        enabled = true,
        icons = { enabled = false },
      },
      dashboard = {
        enabled = true,
        sections = {
          { section = "header" },
          { section = "keys", gap = 1, padding = 1 },
          { section = "recent_files", gap = 1, padding = 1, limit = 8 },
          { section = "startup" },
        },
      },
      notifier = { enabled = true, timeout = 3000 },
      lazygit = { enabled = true },
      terminal = { enabled = true },
      bigfile = { enabled = true },
      indent = { enabled = true },
      scroll = { enabled = true },
      statuscolumn = { enabled = true },
      words = { enabled = true },
    },
  },

  -- ─── UI / Messages (Noice) ────────────────────────────────────────────────
  {
    "folke/noice.nvim",
    event = "VeryLazy",
    opts = {
      lsp = {
        -- override markdown rendering so that cmp and other plugins use Treesitter
        override = {
          ["vim.lsp.util.convert_input_to_markdown_lines"] = true,
          ["vim.lsp.util.stylize_markdown"] = true,
          ["cmp.entry.get_documentation"] = true,
        },
      },
      presets = {
        bottom_search = true,
        command_palette = true,
        long_message_to_split = true,
        inc_rename = false,
        lsp_doc_border = false,
      },
    },
    dependencies = {
      "MunifTanjim/nui.nvim",
    },
  },

  -- ─── Syntax Highlighting ─────────────────────────────────────────────────
  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    -- NOTE: New main branch no longer uses require("nvim-treesitter.configs").setup()
    -- Highlighting is enabled via vim.treesitter natively in Neovim 0.9+
    -- We only need to ensure parsers are installed.
    config = function()
      require("nvim-treesitter").install({
        "bash",
        "javascript",
        "json",
        "lua",
        "markdown",
        "markdown_inline",
        "python",
        "query",
        "toml",
        "vim",
        "vimdoc",
        "yaml",
      })
    end,
  },

  -- ─── Render Markdown ──────────────────────────────────────────────────────
  {
    "MeanderingProgrammer/render-markdown.nvim",
    dependencies = { "nvim-treesitter/nvim-treesitter" },
    opts = {},
    ft = { "markdown", "markdown_inline", "codecompanion" },
  },

  -- ─── LSP ─────────────────────────────────────────────────────────────────
  {
    "neovim/nvim-lspconfig",
    dependencies = {
      "williamboman/mason.nvim",
      "williamboman/mason-lspconfig.nvim",
    },
    config = function()
      require("mason").setup()
      -- mason-lspconfig v2: no more setup_handlers — servers are auto-enabled
      require("mason-lspconfig").setup({
        ensure_installed = {
          "pyright",   -- Python type checking
          "ruff",      -- Python linting/formatting (via LSP)
          "lua_ls",    -- Lua
          "bashls",    -- Bash
          "jsonls",    -- JSON
          "yamlls",    -- YAML
        },
      })

      -- Per-server configuration using the new native API
      local caps = vim.lsp.protocol.make_client_capabilities()
      -- Merge with blink.cmp capabilities (set when blink loads)
      if pcall(require, "blink.cmp") then
        caps = require("blink.cmp").get_lsp_capabilities(caps)
      end

      vim.lsp.config("pyright", { capabilities = caps })
      vim.lsp.config("ruff", { capabilities = caps })
      vim.lsp.config("lua_ls", {
        capabilities = caps,
        settings = {
          Lua = {
            runtime = { version = "LuaJIT" },
            diagnostics = { globals = { "vim", "Snacks" } },
            workspace = { checkThirdParty = false },
          },
        },
      })
      vim.lsp.config("bashls", { capabilities = caps })
      vim.lsp.config("jsonls", { capabilities = caps })
      vim.lsp.config("yamlls", { capabilities = caps })
    end,
  },

  -- ─── Formatting (stevearc/conform) ───────────────────────────────────────
  {
    "stevearc/conform.nvim",
    event = { "BufWritePre" },
    cmd = { "ConformInfo" },
    opts = {
      formatters_by_ft = {
        python    = { "ruff_format", "ruff_organize_imports" },
        lua       = { "stylua" },
        sh        = { "shfmt" },
        bash      = { "shfmt" },
        json      = { "jq" },
        yaml      = { "prettier" },
        markdown  = { "prettier" },
        ["*"]     = { "trim_whitespace" },
      },
      format_on_save = {
        timeout_ms = 3000,
        lsp_fallback = true,
      },
    },
    keys = {
      { "<leader>cf", function() require("conform").format({ async = true, lsp_fallback = true }) end, desc = "Format File" },
    },
  },

  -- ─── Linting (mfussenegger/nvim-lint) ────────────────────────────────────
  {
    "mfussenegger/nvim-lint",
    event = { "BufReadPost", "BufWritePost" },
    config = function()
      local lint = require("lint")
      lint.linters_by_ft = {
        python = { "ruff" },
        sh     = { "shellcheck" },
        bash   = { "shellcheck" },
        yaml   = { "yamllint" },
      }
      vim.api.nvim_create_autocmd({ "BufWritePost", "BufReadPost", "InsertLeave" }, {
        group = vim.api.nvim_create_augroup("nvim_lint", { clear = true }),
        callback = function()
          lint.try_lint()
        end,
      })
    end,
  },

  -- ─── Completion (blink.cmp — replaces nvim-cmp) ──────────────────────────
  {
    "saghen/blink.cmp",
    version = "*",
    opts = {
      keymap = { preset = "default" },
      appearance = {
        use_nvim_cmp_as_default = true,
        nerd_font_variant = "none",
      },
      sources = {
        default = { "lsp", "path", "snippets", "buffer" },
      },
      completion = {
        documentation = { auto_show = true, auto_show_delay_ms = 200 },
        ghost_text = { enabled = true },  -- inline preview like Copilot
      },
    },
  },

  -- ─── Git ─────────────────────────────────────────────────────────────────
  {
    "lewis6991/gitsigns.nvim",
    event = { "BufReadPost", "BufNewFile" },
    opts = {
      signs = {
        add          = { text = "▎" },
        change       = { text = "▎" },
        delete       = { text = "" },
        topdelete    = { text = "" },
        changedelete = { text = "▎" },
      },
      on_attach = function(buffer)
        local gs = package.loaded.gitsigns
        local map = vim.keymap.set
        local opts = { buffer = buffer }
        map("n", "]h", gs.next_hunk, vim.tbl_extend("force", opts, { desc = "Next Hunk" }))
        map("n", "[h", gs.prev_hunk, vim.tbl_extend("force", opts, { desc = "Prev Hunk" }))
        map("n", "<leader>gs", gs.stage_hunk, vim.tbl_extend("force", opts, { desc = "Stage Hunk" }))
        map("n", "<leader>gr", gs.reset_hunk, vim.tbl_extend("force", opts, { desc = "Reset Hunk" }))
        map("n", "<leader>gS", gs.stage_buffer, vim.tbl_extend("force", opts, { desc = "Stage Buffer" }))
        map("n", "<leader>gd", gs.diffthis, vim.tbl_extend("force", opts, { desc = "Diff This" }))
        map("n", "<leader>gb", function() gs.blame_line({ full = true }) end, vim.tbl_extend("force", opts, { desc = "Blame Line" }))
      end,
    },
  },

  -- ─── Which Key ───────────────────────────────────────────────────────────
  {
    "folke/which-key.nvim",
    event = "VeryLazy",
    opts = {
      preset = "modern",
      spec = {
        { "<leader>c", group = "code" },
        { "<leader>g", group = "git" },
        { "<leader>d", group = "document" },
        { "<leader>x", group = "diagnostics" },
      },
    },
  },

  -- ─── Diagnostics Panel ───────────────────────────────────────────────────
  {
    "folke/trouble.nvim",
    cmd = "Trouble",
    keys = {
      { "<leader>xx", "<cmd>Trouble diagnostics toggle<cr>",                        desc = "Diagnostics" },
      { "<leader>xX", "<cmd>Trouble diagnostics toggle filter.buf=0<cr>",           desc = "Buffer Diagnostics" },
      { "<leader>xL", "<cmd>Trouble loclist toggle<cr>",                            desc = "Location List" },
      { "<leader>xQ", "<cmd>Trouble qflist toggle<cr>",                             desc = "Quickfix List" },
    },
    opts = { use_diagnostic_signs = true },
  },

  -- ─── Motion / Navigation ─────────────────────────────────────────────────
  {
    "folke/flash.nvim",
    event = "VeryLazy",
    opts = {},
    keys = {
      { "s",     mode = { "n", "x", "o" }, function() require("flash").jump() end,              desc = "Flash Jump" },
      { "S",     mode = { "n", "x", "o" }, function() require("flash").treesitter() end,        desc = "Flash Treesitter" },
      { "r",     mode = "o",               function() require("flash").remote() end,            desc = "Remote Flash" },
      { "R",     mode = { "o", "x" },      function() require("flash").treesitter_search() end, desc = "Treesitter Search" },
    },
  },

  -- ─── Surround (VSCode vim.surround: true equivalent) ─────────────────────
  {
    "echasnovski/mini.surround",
    version = "*",
    opts = {},
  },

  -- ─── Comments (replaces archived Comment.nvim) ────────────────────────────
  {
    "folke/ts-comments.nvim",
    event = "VeryLazy",
    opts = {},
  },

  -- ─── Python: venv switcher ───────────────────────────────────────────────
  {
    "linux-cultist/venv-selector.nvim",
    branch = "regexp",
    dependencies = { "folke/snacks.nvim" },
    cmd = "VenvSelect",
    keys = {
      { "<leader>pv", "<cmd>VenvSelect<cr>", desc = "Select Python Venv" },
    },
    opts = {
      settings = {
        options = {
          notify_user_on_venv_activation = true,
        },
      },
    },
  },

  -- ─── Legacy text objects (still useful) ──────────────────────────────────
  { "tpope/vim-rsi" },
  { "tpope/vim-unimpaired" },
  { "wellle/targets.vim" },
  { "michaeljsmith/vim-indent-object" },

  -- ─── AI Companion ────────────────────────────────────────────────────────
  {
    "olimorris/codecompanion.nvim",
    dependencies = {
      "nvim-lua/plenary.nvim",
      "nvim-treesitter/nvim-treesitter",
    },
    config = function()
      require("codecompanion").setup({
        strategies = {
          chat   = { adapter = "gemini" },
          inline = { adapter = "gemini" },
        },
      })
      vim.keymap.set({ "n", "v" }, "<C-a>", "<cmd>CodeCompanionActions<cr>",      { noremap = true, silent = true, desc = "AI Actions" })
      vim.keymap.set({ "n", "v" }, "<LocalLeader>a", "<cmd>CodeCompanionChat Toggle<cr>", { noremap = true, silent = true, desc = "AI Chat" })
      vim.keymap.set("v", "ga", "<cmd>CodeCompanionChat Add<cr>",                 { noremap = true, silent = true, desc = "Add to AI Chat" })
    end,
  },
}
