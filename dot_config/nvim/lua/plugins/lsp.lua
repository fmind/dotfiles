-- Configure Language Servers (LSP)

-- Lazily load nvim-lspconfig's shipped gopls config so the root_dir override below can
-- delegate to it. It is only resolvable once the plugin is on the runtimepath, which is
-- later than when this spec is first read.
local gopls_upstream_root_dir
local function upstream_gopls_root_dir()
  if gopls_upstream_root_dir == nil then
    local path = vim.api.nvim_get_runtime_file("lsp/gopls.lua", false)[1]
    local ok, config = pcall(dofile, path or "")
    gopls_upstream_root_dir = (ok and type(config) == "table" and config.root_dir) or false
  end
  return gopls_upstream_root_dir or nil
end

return {
  {
    "neovim/nvim-lspconfig",
    opts = {
      servers = {
        gopls = {
          -- gopls roots at the nearest go.work/go.mod/.git. In a repo that has a go.work
          -- but also carries Go files outside every module (agent skill references such as
          -- `package <slug>`), it attaches at the workspace root and then fails every
          -- request with "no package metadata for file". Require a real module above the
          -- file before attaching; otherwise defer to upstream, which keeps dependency
          -- sources under GOMODCACHE/GOROOT on a single shared server.
          root_dir = function(bufnr, on_dir)
            local fname = vim.api.nvim_buf_get_name(bufnr)
            local go_mod = vim.fs.root(fname, "go.mod")
            if not go_mod then
              return
            end
            local upstream = upstream_gopls_root_dir()
            if upstream then
              return upstream(bufnr, on_dir)
            end
            on_dir(vim.fs.root(fname, "go.work") or go_mod)
          end,
        },
        templ = {},
        tailwindcss = {
          filetypes_include = { "templ" },
          settings = {
            tailwindCSS = {
              includeLanguages = {
                templ = "html",
              },
            },
          },
        },
      },
    },
  },
}
