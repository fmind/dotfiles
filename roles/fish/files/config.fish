# Prompt {{{
function fish_prompt
	if not set -q VIRTUAL_ENV_DISABLE_PROMPT
        set -g VIRTUAL_ENV_DISABLE_PROMPT true
    end

	# Line 1
    set_color yellow
    printf '%s' (whoami)
    set_color normal
    printf ' at '

    set_color magenta
    echo -n (prompt_hostname)
    set_color normal
    printf ' in '

    set_color $fish_color_cwd
    printf '%s' (prompt_pwd)
    set_color normal

    # Line 2
    echo
    if test $VIRTUAL_ENV
        printf "(%s) " (set_color blue)(basename $VIRTUAL_ENV)(set_color normal)
    end
    printf '↪ '
    set_color normal
end
# }}}
# Colors {{{
set fish_color_autosuggestion BD93F9
set fish_color_cancel \x2dr
set fish_color_command F92672
set fish_color_comment 75715E
set fish_color_cwd 66D9EF
set fish_color_cwd_root red
set fish_color_end 50FA7B
set fish_color_error F8F8F2
set fish_color_escape 66D9EF
set fish_color_history_current \x2d\x2dbold
set fish_color_host normal
set fish_color_match F8F8F2
set fish_color_normal F8F8F2
set fish_color_operator AE81FF
set fish_color_param A6E22E
set fish_color_quote E6DB74
set fish_color_redirection AE81FF
set fish_color_search_match \x2d\x2dbackground\x3d49483E
set fish_color_selection white\x1e\x2d\x2dbold\x1e\x2d\x2dbackground\x3dbrblack
set fish_color_status red
set fish_color_user brgreen
set fish_color_valid_path \x2d\x2dunderline
set fish_key_bindings fish_default_key_bindings
set fish_pager_color_completion 75715E
set fish_pager_color_description 49483E
set fish_pager_color_prefix F8F8F2
set fish_pager_color_progress F8F8F2
# }}}
# Plugins {{{
## byobu
status --is-login; and status --is-interactive; and exec byobu-launcher
# }}}
# Options {{{
umask 0077

set -g fish_greeting ''
# }}}
# Exports {{{
set -gx EDITOR vim
# }}}

# review bash and other
# test display in shell
# functions in other file

## EXTERNAL {{{
#if [ -f ~/.environ ]; then
#    . ~/.environ
#fi
#if [ -f ~/.display ]; then
#    . ~/.display
#fi
#if [ -f ~/.aliases ]; then
#    . ~/.aliases
#fi
#if [ -f ~/.private ]; then
#    . ~/.private
#fi
## }}}
