" vim: fdm=marker
" BUFFER {{{
set hidden
set confirm
set autoread
set autowrite
" }}}
" SEARCH {{{
set gdefault
set hlsearch
set incsearch
set smartcase
set ignorecase
" }}}
" INDENT {{{
set expandtab
set shiftround
set tabstop=4
set shiftwidth=4
set softtabstop=4
" }}}
" WINDOW {{{
set linebreak
set lazyredraw
set shortmess=I
set scrolloff=10
" }}}
" SPELL {{{
set spell
set spelllang=en
" }}}
" NUMBER {{{
set number
set relativenumber
" }}}
" FOLDER {{{
set foldmethod=syntax
set foldlevelstart=99
" }}}
" EXTERNAL {{{
set shell=/bin/bash
set clipboard=unnamedplus
" }}}
" COMPLETE {{{
set wildmode=list:longest,full
set completeopt=menuone,longest
" }}}
" PLUGIN {{{
let g:loaded_netrw = 1
let g:loaded_matchparen=1
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'christoomey/vim-tmux-navigator'
Plug 'deoplete-plugins/deoplete-jedi'
Plug 'francoiscabrol/ranger.vim'
let g:ranger_map_keys = 0
Plug 'godlygeek/tabular'
Plug 'honza/vim-snippets'
Plug 'itchyny/lightline.vim'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'rbgrouleff/bclose.vim' " ranger dependency
Plug 'Shougo/deoplete.nvim', {'do': ':UpdateRemotePlugins'}
let g:deoplete#enable_at_startup = 1
Plug 'SirVer/ultisnips'
Plug 'szw/vim-g'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-eunuch'
Plug 'tpope/vim-fugitive'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-unimpaired'
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let g:ale_sign_column_always = 1
let b:ale_fixers = {'python': ['black', 'isort']}
let b:ale_linters = {'python': ['mypy', 'pylint']}
Plug 'wellle/tmux-complete.vim'
call plug#end()
" }}}
" COLOR {{{
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" }}}
" REMAP {{{
noremap j gj
noremap k gk
noremap B g^
noremap E g$
noremap Y y$
noremap U <C-r>
noremap gl :nohl<cr>
inoremap <expr> <C-j> pumvisible() ? "\<C-n>" : "\<C-j>"
inoremap <expr> <C-k> pumvisible() ? "\<C-p>" : "\<C-k>"
xnoremap < <gv
xnoremap > >gv
" }}}
" LEADER {{{
noremap <cr> :
let mapleader=" "
noremap <leader>a :Ag<cr>
noremap <leader>b :Buffers<cr>
noremap <leader>c :Colors<cr>
noremap <leader>d :Tags<cr>
noremap <leader>e :cnext<cr>
noremap <leader>f :Files<cr>
noremap <leader>g :GFiles<cr>
noremap <leader>h :Helptags<cr>
noremap <leader>i :Lines<cr>
noremap <leader>j :bnext<cr>
noremap <leader>k :bprevious<cr>
noremap <leader>l :BLines<cr>
noremap <leader>m :Marks<cr>
noremap <leader>n :BCommits<cr>
noremap <leader>o :call VimuxOpenRunner()<cr>
noremap <leader>p :Commands<cr>
noremap <leader>q :bdelete<cr>:bnext<cr>
noremap <leader>r :Ranger<cr>
noremap <leader>s :Google 
noremap <leader>t :BTags<cr>
noremap <leader>u :VimuxRunLastCommand<cr>
noremap <leader>v "vy :call VimuxSlime(@v)<cr>
noremap <leader>w :Windows<cr>
noremap <leader>x :ALEFix<cr> 
noremap <leader>y :VimuxInspectRunner<cr>
noremap <leader>z :Filetypes<cr>
noremap <leader>` :Locate 
noremap <leader>] :ALENextWrap<cr>
noremap <leader>[ :ALEPreviousWrap<cr>
noremap <leader>' :VimuxPromptCommand<cr>
noremap <leader>; :call VimuxSlime(join(getline(1, '$'), "\n"))<cr>
noremap <leader>. :edit $MYVIMRC<cr>
noremap <leader>, :Gw<cr>
noremap <leader>= :Tabularize 
noremap <leader>: :History:<cr>
noremap <leader>/ :History/<cr>
noremap <leader>\ :History<cr>
noremap <leader>? :Maps<cr>
noremap <leader><cr> :make<cr>
noremap <leader><tab> :b#<cr>
noremap <leader><space> :make 
" }}}
" FUNCTION {{{
function! VimuxSlime(text)
    call VimuxSendText(a:text)
    if a:text !~ '\n$'
        call VimuxSendKeys("Enter")
    endif
endfunction
" }}}
" FILE-TYPES {{{
autocmd FileType yaml setlocal tabstop=2 softtabstop=2 shiftwidth=2
" }}}
" AUTO-COMMANDs {{{
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
" }}}
