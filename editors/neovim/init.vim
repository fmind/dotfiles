" vim: fdm=marker
" STARTING {{{
let $VIMDIR=$HOME.'/.config/nvim'
source ~/.vimrc
"}}}
" PLUGING {{{
syntax enable
filetype plugin on
filetype indent on
if !empty(glob($VIMDIR.'/autoload/plug.vim'))
call plug#begin($VIMDIR.'/plugged')
"" SIDE {{{
Plug 'majutsushi/tagbar'
let g:tagbar_autofocus = 1
nnoremap <leader>j :TagbarToggle<CR>
Plug 'bling/vim-airline' 
let g:airline_powerline_fonts=1
let g:airline#extensions#ale#enabled = 1
let g:airline#extensions#branch#enabled = 1
let g:airline#extensions#tagbar#enabled = 1
let g:airline#extensions#tabline#enabled = 1
let g:airline#extensions#wordcount#enabled = 1
let g:airline#extensions#virtualenv#enabled = 1
Plug 'scrooloose/nerdtree' 
Plug 'Xuyuanp/nerdtree-git-plugin'
Plug 'tiagofumo/vim-nerdtree-syntax-highlight'
let g:NERDTreeQuitOnOpen = 1
let NERDTreeIgnore=['\.pyc$', '\~$']
nnoremap <leader>~ :NERDTreeFind<CR>
nnoremap <leader>` :NERDTreeToggle<CR>
Plug 'airblade/vim-gitgutter'
let g:gitgutter_grep = 'ag'
let g:gitgutter_map_keys = 0
nnoremap ]g <Plug>GitGutterNextHunk
nnoremap [g <Plug>GitGutterPrevHunk
" }}}
"" TARGETS {{{
Plug 'justinmk/vim-sneak'
let g:sneak#label = 1
let g:sneak#s_next = 1
let g:sneak#use_ic_scs = 1
Plug 'wellle/targets.vim'
Plug 'vim-scripts/matchit.zip'
Plug 'christoomey/vim-sort-motion'
Plug 'michaeljsmith/vim-indent-object'
" }}}
"" EDITION {{{
Plug 'tpope/vim-repeat'
Plug 'tpope/vim-abolish'
Plug 'tpope/vim-surround'
Plug 'tpope/vim-commentary'
Plug 'Raimondi/delimitMate'
Plug 'tpope/vim-speeddating'
" }}}
"" INTEGRATION {{{
Plug 'w0rp/ale'
nnoremap <leader>E :ALEToggle<CR>
nnoremap <leader>e :ALENextWrap<CR>
Plug 'tpope/vim-eunuch'
let $FZF_DEFAULT_COMMAND = 'ag --hidden -p ~/.agignore -g ""'
Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
Plug 'junegunn/fzf.vim'
nnoremap <leader>/ :Ag<CR>
nnoremap <leader>A :Tags<CR>
nnoremap <leader>a :BTags<CR>
nnoremap <leader>m :Marks<CR>
nnoremap <leader>L :Lines<CR>
nnoremap <leader>C :Colors<CR>
nnoremap <leader>l :BLines<CR>
nnoremap <leader>F :GFiles<CR>
nnoremap <leader>f :Files .<CR>
nnoremap <leader>b :Buffers<CR>
nnoremap <leader>B :Filetypes<CR>
nnoremap <leader>y :BCommits<CR>
nnoremap <leader>Y :Commits<CR>
nnoremap <leader>w :Windows<CR>
nnoremap <leader>i :Snippets<CR>
nnoremap <leader>h :History<CR>
nnoremap <leader>: :History:<CR>
nnoremap <leader>; :History/<CR>
nnoremap <leader>? :Helptags<CR>
nnoremap <leader><Space> :Commands<CR>
Plug 'SirVer/ultisnips'
nnoremap <leader>I :UltiSnipsEdit<CR>
let g:UltiSnipsEditSplit = 'context'
let g:UltiSnipsExpandTrigger="<tab>"
let g:UltiSnipsListSnippets="<s-tab>"
let g:UltiSnipsJumpForwardTrigger="<c-n>"
let g:UltiSnipsJumpBackwardTrigger="<c-p>"
let g:UltiSnipsSnippetsDir = $VIMDIR.'/snippets/'
Plug 'tpope/vim-fugitive'
noremap <Leader>G Git 
noremap <Leader>gj :Glcd 
noremap <Leader>gh :Gpush<CR>
noremap <Leader>gl :Gpull<CR>
noremap <Leader>gm :Gmove
noremap <Leader>gw :Gwrite<CR>
noremap <Leader>gc :Gcommit<CR>
noremap <Leader>gs :Gstatus<CR>
noremap <Leader>go :Gbrowse<CR>
noremap <Leader>gb :Gblame<CR>
noremap <Leader>gd :Gvdiff<CR>
noremap <Leader>gr :Gremove<CR>
Plug 'Shougo/deoplete.nvim', { 'do': ':UpdateRemotePlugins', 'for': 'python' }
Plug 'zchee/deoplete-jedi'
Plug 'wellle/tmux-complete.vim'
let g:deoplete#enable_at_startup = 1
Plug 'christoomey/vim-tmux-navigator'
let g:tmux_navigator_save_on_switch = 1
Plug 'beloglazov/vim-online-thesaurus'
let g:online_thesaurus_map_keys = 0
nnoremap <leader>U :Thesaurus 
nnoremap <leader>u :OnlineThesaurusCurrentWord<CR>
" }}}
"" THEME {{{
Plug 'tomasr/molokai'
colorscheme molokai
let g:molokai_original = 1
" }}}
"
"" BASE {{{
" denite?
" pencil ?
" DEVICON
" vim session
"
"let g:session_directory = "~/.vim/session"
"let g:session_autoload = "no"
"let g:session_autosave = "no"
"let g:session_command_aliases = 1
"nnoremap <leader>so :OpenSession<CR>
"nnoremap <leader>ss :SaveSession<CR>
"nnoremap <leader>sd :DeleteSession<CR>
"nnoremap <leader>sc :CloseSession<CR>
" }}}


""" TUNING {{{
" Plug 'moll/vim-bbye'
" Plug 'benmills/vimux'
" Plug 'junegunn/vim-easy-align'
"}}}
"""" WRITING {{{
"Plug 'godlygeek/tabular'
"Plug 'plasticboy/vim-markdown'
""}}}
"""" DEVELOPPING {{{
"" Plug 'janko-m/vim-test'
"" let g:test#preserve_screen = 1
"" let test#strategy = "vimux"
"" let test#python#runner = 'pytest'
"
"Plug 'sheerun/vim-polyglot'
"" Plug 'aklt/plantuml-syntax'
"

""}}}
call plug#end()
endif
"""}}}
""}}}
" BINDING {{{
"" ACTIONS {{{
"}}}

"""" PLUGINS CONFIGURATIONS {{{
""""""" edition {{{
"xmap ga <Plug>(EasyAlign)
"nmap ga <Plug>(EasyAlign)
"nnoremap <leader>A :Tabularize
"nnoremap <leader>a= :Tabularize /=
"vnoremap <leader>a= :Tabularize /=
"nnoremap <leader>a/ :Tabularize /|
"vnoremap <leader>a/ :Tabularize /|
"nnoremap <leader>q :Autoformat<CR>
""}}}
""""""" testing {{{
"" nnoremap <leader>rf :TestFile<CR>
"" nnoremap <leader>rl :TestLast<CR>
"" nnoremap <leader>rs :TestSuite<CR>
"" nnoremap <leader>rv :TestVisit<CR>
"" nnoremap <leader>rr :TesgNearest<CR>
""}}}
""""""" toggling {{{
"" nnoremap <leader>G :HardPencil<CR>
""}}}
""""""" tmux management {{{
"nnoremap <Leader>vq :VimuxCloseRunner<CR>
"nnoremap <Leader>vv :VimuxPromptCommand<CR>
"nnoremap <Leader>vi :VimuxInspectRunner<CR>
"nnoremap <Leader>vl :VimuxRunLastCommand<CR>
"nnoremap <Leader>vx :VimuxInterruptRunner<CR>
"nnoremap <Leader>vz :call VimuxZoomRunner()<CR>
""}}}
""""""" error management {{{
""}}}
""""""" text management {{{
"autocmd! BufNewFile,BufRead *.{txt,md} setlocal spell spelllang=en,fr whichwrap+=h,l,<,>,[,]
""}}}
""""""" uml language {{{
"" autocmd BufWritePost *.uml :silent !plantuml <afile>
""}}}
""""""" python language {{{
"" vim-virtualenv
"" autocmd BufWritePost *.py :Isort
"" autocmd BufWritePost *.py :Autoformat
"" autocmd FileType python setlocal foldmethod=indent
"" autocmd FileType python nnoremap <localleader>i :Isort<CR>
""}}}
""}}}
