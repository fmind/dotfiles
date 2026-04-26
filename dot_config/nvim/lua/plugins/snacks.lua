return {
	{
		"folke/snacks.nvim",
		opts = {
			image = { enabled = false },
			picker = {
				sources = {
					explorer = {
						actions = {
							explorer_open_all = function(picker)
								local Tree = require("snacks.explorer.tree")
								local function recurse(node)
									if not node.dir then
										return
									end
									Tree:expand(node)
									node.open = true
									for _, child in pairs(node.children) do
										recurse(child)
									end
								end
								local root = Tree:node(picker:cwd())
								if root then
									recurse(root)
									picker:find()
								end
							end,
						},
						win = {
							list = {
								keys = {
									["O"] = "explorer_open_all",
								},
							},
						},
					},
				},
			},
		},
	},
}
