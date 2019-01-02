" vim: fdm=marker
" INIT {{{
set runtimepath^=~/.vim runtimepath+=~/.vim/after
let g:python3_host_prog = '/usr/bin/python3'
let &packpath=&runtimepath
source ~/.vimrc
" }}}
" PLUGIN {{{
call plug#begin('~/.local/share/nvim/plugged')
Plug 'w0rp/ale'
let g:ale_set_loclist = 0
let g:ale_set_quickfix = 1
let g:ale_sign_column_always = 1
call plug#end()
" }}}
" KEYMAPS {{{
nnoremap gk :terminal<CR>
tnoremap <Esc> <C-\><C-n>
" }}}
" AUTOCMD {{{
augroup nterm
    autocmd!
    autocmd TermOpen * startinsert
augroup end
" }}}
