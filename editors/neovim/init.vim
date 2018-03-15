" vim: fdm=marker
" STARTING {{{
let $VIMDIR=$HOME.'/.config/nvim'
source ~/.vimrc
"}}}
" SETTING {{{
set relativenumber
"}}}
" PLUGING {{{
syntax enable
filetype plugin on
filetype indent on
if !empty(glob($VIMDIR.'/autoload/plug.vim'))
call plug#begin($VIMDIR.'/plugged')
"" SIDE {{{
Plug 'scrooloose/nerdtree' 
let g:NERDTreeQuitOnOpen = 1
let NERDTreeIgnore=['\.pyc$', '\~$']
nnoremap <silent> <leader>` :NERDTreeToggle<CR>
Plug 'Xuyuanp/nerdtree-git-plugin'
Plug 'tiagofumo/vim-nerdtree-syntax-highlight'
" Plug 'majutsushi/tagbar'
Plug 'bling/vim-airline' 
" test powerline fonts symbol >
let g:airline_powerline_fonts=1
let g:airline#extensions#tabline#enabled = 1
Plug 'airblade/vim-gitgutter'
let g:gitgutter_grep = 'ag'
let g:gitgutter_map_keys = 0
nnoremap ]c <Plug>GitGutterNextHunk
nnoremap [c <Plug>GitGutterPrevHunk
nnoremap <silent> <leader>G :GitGutterToggle<CR>
" }}}
"" THEME {{{
Plug 'tomasr/molokai'
colorscheme molokai
let g:molokai_original = 1
" }}}
"" BASE {{{
" Plug 'tpope/vim-repeat'
" Plug 'wellle/targets.vim'
" denite?
" Plug 'tpope/vim-surround'
" Plug 'tpope/vim-commentary'
" Plug 'Raimondi/delimitMate'
" Plug 'vim-scripts/matchit.zip'
" Plug 'christoomey/vim-tmux-navigator'
" let g:tmux_navigator_save_on_switch = 1
" }}}

"Plug 'SirVer/ultisnips'
"let g:UltiSnipsEditSplit = 'context'
"let g:UltiSnipsSnippetsDir = $HOME.'/.vim/snippets/'

""" TUNING {{{
" Plug 'tpope/vim-rsi'
" Plug 'moll/vim-bbye'
" Plug 'benmills/vimux'
" Plug 'tpope/vim-eunuch'
" Plug 'justinmk/vim-sneak'
" let g:sneak#label = 1
" let g:sneak#s_next = 1
" let g:sneak#use_ic_scs = 1
" Plug 'alvan/vim-closetag'
" Plug 'tpope/vim-speeddating'
" Plug 'junegunn/vim-easy-align'
" Plug 'christoomey/vim-sort-motion'
" Plug 'michaeljsmith/vim-indent-object'
"}}}
"""" WRITING {{{
"Plug 'godlygeek/tabular'
"Plug 'plasticboy/vim-markdown'
"Plug 'beloglazov/vim-online-thesaurus'
""}}}
"""" DEVELOPPING {{{
"Plug 'w0rp/ale'
"let $FZF_DEFAULT_COMMAND = 'ag -p ~/.agignore -g ""'
"Plug 'junegunn/fzf', {'dir': '~/.fzf', 'do': './install --bin'}
"Plug 'junegunn/fzf.vim'
" let $FZF_DEFAULT_COMMAND = 'ag -p ~/.agignore -g ""'
"" Plug 'janko-m/vim-test'
"" let g:test#preserve_screen = 1
"" let test#python#runner = 'pytest'
"Plug 'sheerun/vim-polyglot'
"" Plug 'aklt/plantuml-syntax'
"

" Plug 'ryanoasis/vim-devicons'
""}}}
call plug#end()
endif
"""}}}
""}}}
" BINDING {{{
"" ACTIONS {{{
"}}}

"vnoremap <leader>s :sort<CR>
"nnoremap <leader>oj :join<CR>
"vnoremap <leader>oj :join<CR>
"""" PLUGINS CONFIGURATIONS {{{
""""""" edition {{{
"xmap ga <Plug>(EasyAlign)
"nmap ga <Plug>(EasyAlign)
"nnoremap <silent> <leader>A :Tabularize
"nnoremap <silent> <leader>a= :Tabularize /=
"vnoremap <silent> <leader>a= :Tabularize /=
"nnoremap <silent> <leader>a/ :Tabularize /|
"vnoremap <silent> <leader>a/ :Tabularize /|
"nnoremap <silent> <leader>q :Autoformat<CR>
""}}}
""""""" navigation {{{
"nnoremap <silent> <leader>/ :Ag<CR>
"nnoremap <silent> <leader>T :Tags<CR>
"nnoremap <silent> <leader>t :BTags<CR>
"nnoremap <silent> <leader>' :Marks<CR>
"nnoremap <silent> <leader>L :Lines<CR>
"nnoremap <silent> <leader>l :BLines<CR>
"nnoremap <silent> <leader>F :GFiles<CR>
"nnoremap <silent> <leader>f :Files .<CR>
"nnoremap <silent> <leader>B :Buffers<CR>
"nnoremap <silent> <leader>y :History<CR>
"nnoremap <silent> <leader>w :Windows<CR>
"nnoremap <silent> <leader>i :Snippets<CR>
"nnoremap <silent> <leader><Space> :Commands<CR>
"nnoremap <silent> <leader>d :YcmCompleter GetDoc<CR>
"nnoremap <silent> <leader>g :YcmCompleter GoToDefinitionElseDeclaration<CR>
""}}}
""""""" definition {{{
"nnoremap <leader>U :Thesaurus 
"nnoremap <leader>u :OnlineThesaurusCurrentWord<CR>
""}}}
""""""" testing {{{
"" nnoremap <silent> <leader>rf :TestFile<CR>
"" nnoremap <silent> <leader>rl :TestLast<CR>
"" nnoremap <silent> <leader>rs :TestSuite<CR>
"" nnoremap <silent> <leader>rv :TestVisit<CR>
"" nnoremap <silent> <leader>rr :TesgNearest<CR>
""}}}
""""""" toggling {{{
"" nnoremap <silent> <leader>G :HardPencil<CR>
"nnoremap <silent> <leader>; :TagbarToggle<CR>
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
"nnoremap <silent> <leader>E :ALEToggle<CR>
"nnoremap <silent> <leader>e :ALENextWrap<CR>
""}}}
""""""" text management {{{
"autocmd! BufNewFile,BufRead *.{txt,md} setlocal spell spelllang=en,fr whichwrap+=h,l,<,>,[,]
""}}}
""""""" session management {{{
"nnoremap <silent> <leader>NL :SLoad<CR>
"nnoremap <silent> <leader>NS :SSave<CR>
"nnoremap <silent> <leader>NC :SClose<CR>
"nnoremap <silent> <leader>ND :SDelete<CR>
""}}}
""""""" snippet management {{{
"nnoremap <silent> <leader>I :UltiSnipsEdit<CR>
""}}}
""""""" uml language {{{
"" autocmd BufWritePost *.uml :silent !plantuml <afile>
""}}}
""""""" python language {{{
"" autocmd BufWritePost *.py :Isort
"" autocmd BufWritePost *.py :Autoformat
"" autocmd FileType python setlocal foldmethod=indent
"" autocmd FileType python nnoremap <silent> <localleader>i :Isort<CR>
"" autocmd FileType python nnoremap <silent> <localleader>t :SwitchPyTest<CR>
""}}}
""}}}
