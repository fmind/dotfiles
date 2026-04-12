# Base commands
set -gx FZF_DEFAULT_COMMAND 'fd -t f -H -E .git'
set -gx FZF_ALT_C_COMMAND 'fd -t d -H -E .git'

# Default options
set -gx FZF_DEFAULT_OPTS '--height=50% --layout=reverse --border --info=inline '\
'--color=bg+:#313244,bg:#1e1e2e,spinner:#f5e0dc,hl:#f38ba8 '\
'--color=fg:#cdd6f4,header:#f38ba8,info:#cba6f7,pointer:#f5e0dc '\
'--color=marker:#b4befe,fg+:#cdd6f4,prompt:#cba6f7,hl+:#f38ba8 '\
'--color=selected-bg:#45475a,border:#313244,label:#cdd6f4'

# Preview options
set -gx FZF_ALT_C_OPTS "--preview 'eza --tree --color=always {} | head -200'"
set -gx FZF_CTRL_T_OPTS "--preview 'bat -n --color=always --line-range=:500 {}'"
