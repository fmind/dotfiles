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
call plug#end()
" }}}
" KEYMAPS {{{
" other {{{
nnoremap gk :terminal<CR>
tnoremap <Esc> <C-\><C-n>
" }}}
" leader {{{
let mapleader="\<space>"
nnoremap <leader><Space> :mak<CR>
" }}}
" localleader {{{
let maplocalleader="\<c-space>"
" }}}
" }}}
" AUTOCMD {{{
augroup nterm
    autocmd!
    autocmd TermOpen * startinsert
augroup end
" }}}
