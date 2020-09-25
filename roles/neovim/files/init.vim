" vim: fdm=marker
" BUFFER {{{
set hidden
set confirm
set autoread
set autowrite
set splitright
set splitbelow
" }}}
" FOLDER {{{
set foldlevelstart=99
" }}}
" FORMAT {{{
set formatoptions-=cro
" }}}
" INDENT {{{
set tabstop=4
set expandtab
set shiftround
set shiftwidth=4
set softtabstop=4
" }}}
" NUMBER {{{
set number
set relativenumber
" }}}
" OTHERS {{{
set shell=/bin/bash
set clipboard=unnamedplus
let s:windows_clip = '/mnt/c/Windows/System32/clip.exe'
" }}}
" POPUPS {{{
set wildmode=list:longest,full
set completeopt=menuone,longest
" }}}
" SEARCH {{{
set gdefault
set hlsearch
set incsearch
set smartcase
set ignorecase
" }}}
" SPELLS {{{
set spell
set spelllang=en,fr
" }}}
" WINDOW {{{
set linebreak
set lazyredraw
set shortmess=I
set scrolloff=10
" }}}
" PLUGIN {{{
let g:loaded_matchparen=1
call plug#begin('~/.local/share/nvim/plugged')
Plug 'benmills/vimux'
Plug 'davidhalter/jedi-vim'
let g:jedi#auto_vim_configuration = 0
Plug 'farmergreg/vim-lastplace'
Plug 'francoiscabrol/ranger.vim'
let g:ranger_map_keys = 0
let g:ranger_replace_netrw = 1
Plug 'godlygeek/tabular'
Plug 'honza/vim-snippets'
Plug 'itchyny/lightline.vim'
Plug 'janko/vim-test'
let test#python#runner = "pytest"
let g:test#preserve_screen = 1
Plug 'jiangmiao/auto-pairs'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
Plug 'justinmk/vim-sneak' 
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'liuchengxu/vim-which-key'
Plug 'majutsushi/tagbar'
Plug 'michaeljsmith/vim-indent-object'
Plug 'rbgrouleff/bclose.vim' " ranger dependency
Plug 'sheerun/vim-polyglot'
Plug 'SirVer/ultisnips'
Plug 'szw/vim-g'
Plug 'tomasr/molokai'
Plug 'tpope/vim-commentary'
Plug 'tpope/vim-eunuch'
Plug 'tpope/vim-fugitive'
Plug 'tpope/vim-projectionist'
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-rsi'
Plug 'tpope/vim-speeddating'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-unimpaired'
Plug 'w0rp/ale'
let g:ale_set_quickfix = 1
let b:ale_fixers = {'python': ['black', 'isort']}
let b:ale_linters = {'python': ['mypy', 'pylint']}
let g:ale_python_pylint_options = '--error-only'
call plug#end()
" }}}
" COLORS {{{
try
    colorscheme molokai
catch
    colorscheme zellner
endtry
" }}}
" REMAPS {{{
nnoremap j gj
nnoremap k gk
nnoremap B g^
nnoremap E g$
nnoremap Y y$
xnoremap < <gv
xnoremap > >gv
nnoremap U <C-r>
nnoremap gl :nohl<CR>
" }}}
" LEADERS {{{
noremap <CR> :
let mapleader="\<space>"
nnoremap <silent> <leader> :WhichKey '<Space>'<CR>
vnoremap <silent> <leader> :WhichKeyVisual '<Space>'<CR>
noremap <leader>a :A<CR>
noremap <leader>b :Buffers<CR>
noremap <leader>c :Colors<CR>
noremap <leader>d :Ag<CR>
noremap <leader>e :Files<CR>
noremap <leader>f :GFiles<CR>
noremap <leader>g :Google 
noremap <leader>h :Helptags<CR>
noremap <leader>i :Lines<CR>
noremap <leader>j :bnext<CR>
noremap <leader>k :bprevious<CR>
noremap <leader>l :BLines<CR>
noremap <leader>m :Marks<CR>
noremap <leader>n :BCommits<CR>
noremap <leader>o :call VimuxOpenRunner()<CR>
noremap <leader>p :Commands<CR>
noremap <leader>q :bdelete<CR>:bnext<CR>
noremap <leader>r :VimuxRunLastCommand<CR>
noremap <leader>s :Snippets<CR>
noremap <leader>t :BTags<CR>
noremap <leader>u :UltiSnipsEdit<CR>
noremap <leader>v "vy :call VimuxSlime(@v)<CR>
noremap <leader>w :Windows<CR>
noremap <leader>x :ALEFix<CR>
noremap <leader>y :Filetypes<CR>
noremap <leader>z :VimuxZoomRunner<CR>
noremap <leader><CR> :make<CR>
noremap <leader><tab> :b#<CR>
noremap <leader><space> :make
noremap <leader>< :prev<CR>
noremap <leader>> :lnext<CR>
noremap <leader>( :cprev<CR>
noremap <leader>) :cnext<CR>
noremap <leader>} :tnext<CR>
noremap <leader>{ :tprev<CR>
noremap <leader>] :ALENextWrap<CR>
noremap <leader>[ :ALEPreviousWrap<CR>
noremap <leader>' :VimuxPromptCommand<CR>
noremap <leader>" :VimuxInspectRunner<CR>
noremap <leader>; :call VimuxSlime(join(getline(1, '$'), "\n"))<CR>
noremap <leader>~ :RangerWorkingDirectory<CR>
noremap <leader>` :RangerCurrentFile<CR>
noremap <leader>- :Locate 
noremap <leader>= :Tabularize 
noremap <leader>_ :GFiles?<CR>
noremap <leader>+ :Commits<CR>
noremap <leader>@ :RainbowToggle<CR>
noremap <leader># :ALEToggle<CR>
noremap <leader>$ :TagbarToggle<CR>
noremap <leader>. :edit $MYVIMRC<CR>
noremap <leader>, :Gw<CR>
noremap <leader>? :Maps<CR>
noremap <leader>\| :Tags<CR>
noremap <leader>\ :History<CR>
noremap <leader>: :History:<CR>
noremap <leader>/ :History/<CR>
" }}}
" LLOCALS {{{
let maplocalleader = ";"
nnoremap <silent> <localleader> :WhichKey ';'<CR>
vnoremap <silent> <localleader> :WhichKeyVisual ';'<CR>
" jedi {{{
noremap <localleader>ji :Pyimport 
let g:jedi#completions_command = "<C-Space>"
let g:jedi#documentation_command = "<localleader>jh"
let g:jedi#goto_assignments_command = "<localleader>ja"
let g:jedi#goto_command = "<localleader>jj"
let g:jedi#goto_definitions_command = "<localleader>jd"
let g:jedi#goto_stubs_command = "<localleader>js"
let g:jedi#rename_command = "<localleader>jr"
let g:jedi#usages_command = "<localleader>ju"
" }}}
" files {{{
noremap <localleader>ec :e .coveragerc<CR>
noremap <localleader>ei :e .gitignore<CR>
noremap <localleader>el :e LICENSE.txt<CR>
noremap <localleader>em :e mypy.ini<CR>
noremap <localleader>er :e README.md<CR>
noremap <localleader>er :e requirements.txt<CR>
noremap <localleader>es :e setup.py<CR>
noremap <localleader>et :e pytest.ini<CR>
noremap <localleader>et :e tasks.py<CR>
noremap <localleader>ey :e pylintrc<CR>
noremap <localleader>eA :e ~/.agignore<CR>
noremap <localleader>eB :e ~/.bashrc<CR>
noremap <localleader>eC :e ~/.cookiecutterrc<CR>
noremap <localleader>eG :e ~/.gitconfig<CR>
noremap <localleader>eI :e ~/.gitignore<CR>
noremap <localleader>eJ :e ~/.jupyter/.jupyter_notebook_config.py<CR>
noremap <localleader>eO :e ~/.condarc<CR>
noremap <localleader>eP :e ~/.ipython/profile_default/ipython_config.py<CR>
noremap <localleader>eS :e ~/.ssh/config<CR>
noremap <localleader>eT :e ~/.ctags<CR>
noremap <localleader>eV :e ~/.config/nvim/init.vim<CR>
noremap <localleader>eX :e ~/.xonshrc<CR>
noremap <localleader>eY :e ~/.pypirc<CR>
" }}}
" plugs {{{
noremap <localleader>xd :PlugDiff<CR>
noremap <localleader>xc :PlugClean<CR>
noremap <localleader>xi :PlugInstall<CR>
noremap <localleader>xu :PlugUpdate<CR>
noremap <localleader>xg :PlugUpgrade<CR>
noremap <localleader>xw :PlugSnapshot<CR>
noremap <localleader>xs :PlugStatus<CR>
" }}}
" tests {{{
noremap <localleader>tf :TestFile<CR>
noremap <localleader>tl :TestLast<CR>
noremap <localleader>ts :TestSuite<CR>
noremap <localleader>tt :TestNearest<CR>
noremap <localleader>tv :TestVisit<CR>
" }}}
" spells {{{
noremap <localleader>la :set spelllang=en,fr<CR>
noremap <localleader>le :set spelllang=en<CR>
noremap <localleader>lf :set spelllang=fr<CR>
noremap <localleader>ls :set nospell<CR>
" }}}
" python {{{
noremap <localleader>pb :!bandit %<CR>
noremap <localleader>pc :!coverage %<CR>
noremap <localleader>pd :!pydoc3 
noremap <localleader>pe :!python3 %<CR>
noremap <localleader>pf :!black %<CR>
noremap <localleader>pi :!isort %<CR>
noremap <localleader>pl :!pylint %<CR>
noremap <localleader>pm :!mypy %<CR>
noremap <localleader>po :!inv
noremap <localleader>pr :!rope
noremap <localleader>pt :!pytest %<CR>
noremap <localleader>pu :!vulture %<CR>
noremap <localleader>py :!ipython -i %<CR>
noremap <localleader>pvv :!python3 -m venv venv<CR>
noremap <localleader>ppi :!python3 -m pip install 
noremap <localleader>ppn :!python3 -m pip install pynvim<CR>
" }}}
" }}}
" AUTO-GROUPS {{{
" AutoVim {{{
augroup AutoVim
    autocmd!
    autocmd BufWritePost $MYVIMRC source $MYVIMRC
augroup END
" }}}
" WSLYank {{{
if executable(s:windows_clip)
    augroup WSLYank
        autocmd!
        autocmd TextYankPost * if v:event.operator ==# 'y' | call system(s:windows_clip, @0) | endif
    augroup END
endif
" }}}
" }}}
