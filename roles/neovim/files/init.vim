" CLIP
set clipboard=unnamedplus
" GREP
if executable('ag')
  set grepprg=ag\ --nogroup\ --nocolor
endif
" MENU
set wildmode=list:longest,full
set completeopt=menuone,longest
" SHELL
set shell=/bin/bash
" SPELL
set spell
set spelllang=en
set dictionary+=/usr/share/dict/words
" BUFFER
set hidden
set confirm
set autoread
set autowrite
" SEARCH
set path+=**
set gdefault
set hlsearch
set incsearch
set smartcase
set ignorecase
" INDENT
set expandtab
set shiftround
set shiftwidth=4
" WINDOW
set linebreak
set shortmess=I
set scrolloff=10
" NUMBER
set number
set relativenumber
" FOLDER
set foldmethod=syntax
set foldlevelstart=99
" PLUGIN
let g:loaded_netrw = 1
let g:loaded_matchparen=1
let g:loaded_netrwPlugin = 1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'itchyny/lightline.vim'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'Shougo/deoplete.nvim', { 'do': ':UpdateRemotePlugins' }
let g:deoplete#enable_at_startup = 1
Plug 'deoplete-plugins/deoplete-jedi'
Plug 'wellle/tmux-complete.vim'
Plug 'szw/vim-g'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-surround'
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let g:ale_sign_column_always = 1
let b:ale_fixers = {'python': ['black', 'isort']}
let b:ale_linters = {'python': ['mypy', 'pylint']}
call plug#end()
" COLOR
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" REMAP
noremap j gj
noremap k gk
noremap B g^
noremap E g$
noremap Y y$
noremap U <C-r>
xnoremap < <gv
xnoremap > >gv
noremap gl :nohl<cr>
" LEADER
noremap <cr> :
let mapleader=" "
let maplocalleader=";"
noremap <leader>a :Ag<cr>
noremap <leader>b :Buffers<cr>
noremap <leader>c :Colors<cr>
noremap <leader>d :VimuxPromptCommand<cr>
noremap <leader>e "vy :call VimuxSlime(@v)<cr>
noremap <leader>f :Files<cr>
noremap <leader>g :GFiles<cr>
noremap <leader>h :cprevious<cr>
noremap <leader>i :Lines<cr>
noremap <leader>j :bnext<cr>
noremap <leader>k :bprevious<cr>
noremap <leader>l :cnext<cr>
noremap <leader>m :Marks<cr>
noremap <leader>n :BCommits<cr>
noremap <leader>o :Google
noremap <leader>p :Commands<cr>
noremap <leader>q :bdelete<cr>
noremap <leader>r :History<cr>
noremap <leader>s :Tags<cr>
noremap <leader>t :call VimuxOpenRunner()<cr>
noremap <leader>u :VimuxRunLastCommand<cr>
noremap <leader>v :BTags<cr>
noremap <leader>w :Windows<cr>
noremap <leader>x :History:<cr>
noremap <leader>y :VimuxInterruptRunner<cr>
noremap <leader>z :Filetypes<cr>
noremap <leader>` :Locate 
noremap <leader>- :Maps<cr>
noremap <leader>= :ALEFix<cr>
noremap <leader>[ :ALEPreviousWrap<cr>
noremap <leader>] :ALENextWrap<cr>
noremap <leader>' :Helptags<cr>
noremap <leader>, :edit $MYVIMRC<cr>
noremap <leader>. :call VimuxSlime(join(getline(1, '$'), "\n"))<cr>
noremap <leader>/ :History/<cr>
noremap <leader>\ :BLines<cr>
noremap <leader><cr> :make<cr>
noremap <leader><tab> :b#<cr>
noremap <leader><space> :make 
" AUTOCMD
augroup vim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup end
" FUNCTION
function! VimuxSlime(text)
    call VimuxSendText(a:text)
    if a:text !~ '\n$'
        call VimuxSendKeys("Enter")
    endif
endfunction
