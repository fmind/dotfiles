" INIT {{{
source ~/.vimrc
" }}}
" CONFIG {{{
let g:python3_host_prog = '/usr/bin/python3'
" }}}
" PLUGIN {{{
call plug#begin('~/.local/share/nvim/plugged')
Plug 'morhetz/gruvbox'
call plug#end()
" }}}
" COLORS {{{
colorscheme gruvbox
" }}}
" KEYMAPS {{{
" command {{{
nnoremap g, :edit ~/.vimrc<CR>
" }}}
" }}}
