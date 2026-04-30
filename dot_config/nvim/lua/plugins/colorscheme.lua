return {
	{
		"catppuccin/nvim",
		name = "catppuccin",
		opts = {
			flavour = "mocha",
			integrations = {
				aerial = true,
				cmp = true,
				gitsigns = true,
				mason = true,
				neotree = true,
				noice = true,
				notify = true,
				snacks = true,
				treesitter = true,
				which_key = true,
			},
		},
	},
	{
		"LazyVim/LazyVim",
		opts = {
			colorscheme = "catppuccin-mocha",
		},
	},
}
