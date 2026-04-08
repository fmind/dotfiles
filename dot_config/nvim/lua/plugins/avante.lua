return {
  {
    "yetone/avante.nvim",
    opts = {
      provider = "gemini-cli",
      mode = "agentic",
      acp_providers = {
        ["gemini-cli"] = {
          command = "gemini",
          args = { "--acp" },
          env = {
            NODE_NO_WARNINGS = "1",
          },
        },
      },
    },
  },
}
