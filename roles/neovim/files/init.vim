" INIT
set shortmess=I
" PATH
set path+=**
" SHELL
set shell=/bin/bash
" XCLIP
set clipboard=unnamedplus
" MENUS
set wildmode=list:longest,full
set completeopt=menuone,preview
" INDENT
set expandtab
" BUFFER
set hidden
set confirm
" NUMBER
set number
set relativenumber
" WINDOW
set scrolloff=10
set statusline=\ %n:\ \%f\ %y%=%r\ %l\ :\ %c\ (%p%%)\
" COLORS
colorscheme zellner
" PLUGINS
let g:loaded_netrw = 1
let g:loaded_matchparen=0
let g:loaded_netrwPlugin = 1
" KEYMAPS
noremap <CR> :
" LEADERS
let mapleader=" "
noremap <leader>l :nohl<CR>
noremap <leader>. :edit $MYVIMRC<CR>
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
