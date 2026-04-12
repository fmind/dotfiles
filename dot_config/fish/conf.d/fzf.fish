# Base commands
set -gx FZF_DEFAULT_COMMAND 'fd -t f -H -E .git'
set -gx FZF_ALT_C_COMMAND 'fd -t d -H -E .git'

# Default options
set -gx FZF_DEFAULT_OPTS '--height=50% --layout=reverse --border --info=inline'

# Preview options
set -gx FZF_ALT_C_OPTS "--preview 'eza --tree --color=always {} | head -200'"
set -gx FZF_CTRL_T_OPTS "--preview 'bat -n --color=always --line-range=:500 {}'"
