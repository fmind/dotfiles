# a:delta
alias a="delta"
# b:bat
alias b="bat"
# c:gcloud
alias c="gcloud"
# d:docker
alias d="docker"
alias ld="lazydocker"
# e:chezmoi
alias e="chezmoi --source '~/.dotfiles'"
# f:fd
alias f="fd"
# g:git
alias g="git"
alias lg="lazygit"
alias gcop="copilot"
# h:gh
alias h="gh"
# i:gemini
alias i="gemini"
# j:zellij
alias j="zellij"
# k:kubectl
alias k="kubectl"
alias k9="k9s"
# l:eza (ls)
alias l="eza -la"
# m:mise
alias m="mise"
# n:npm
alias n="npm"
# o:btop
alias o="btop"
# p:python3
alias p="python3"
alias pt="ptpython"
# q:gemini (prompt)
alias q="gemini --prompt"
# r:ripgrep (grep)
alias r="rg"
# s:ssh
alias s="ssh"
# t:terraform
alias t="terraform"
# u:uv
alias u="uv"
# v:nvim (vim)
alias v="nvim"
# w:xh (web)
alias w="xh"
# x:open
if command -v xdg-open >/dev/null
	alias x="xdg-open"
else if command -v open >/dev/null
	alias x="open"
end
# y:fzf
alias y="fzf"
# z:zoxide
# alias z=""
