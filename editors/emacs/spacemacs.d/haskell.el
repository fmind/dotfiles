(require 'intero)

                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'haskell-mode
  "-" 'hindent-reformat-decl
  "=" 'haskell-mode-stylish-buffer
  "'" 'haskell-intero/pop-to-repl
  "\"" 'haskell-intero/display-repl
  "," 'intero-repl-load
  "\/" 'helm-hoogle
  ;; "A"
  "B" 'intero-list-buffers
  ;; "C"
  "D" 'intero-devel-reload
  ;; "E"
  ;; "F" RESERVED -- stylish
  ;; "G"
  ;; "H"
  ;; "I"
  ;; "J"
  ;; "K"
  ;; "L"
  ;; "M"
  ;; "N"
  ;; "O"
  ;; "P"
  "Q" 'intero-destroy
  "R" 'intero-restart
  ;; "S"
  ;; "T"
  ;; "U"
  ;; "V"
  ;; "W"
  ;; "X"
  ;; "Y"
  ;; "Z"
  "a" 'intero-apply-suggestions
  "b" 'hlint-refactor-refactor-buffer
  ;; "c" RESERVED -- cabal
  ;; "d" RESERVED -- debug
  "e" 'intero-repl-eval-region
  ;; "f" RESERVED -- hindent
  ;; "g" RESERVED -- navigation
  ;; "h" RESERVED -- documentation
  ;; "i" RESERVED -- intero
  "j" 'intero-goto-definition
  "k" 'hoogle
  "l" 'intero-targets
  "m" 'intero-expand-splice-at-point
  "n" 'haskell-navigate-imports
  "o" 'intero-info
  "p" 'hlint-refactor-refactor-at-point
  ;; "q"
  ;; "r" RESERVED -- refactor
  ;; "s" RESERVED -- repl
  "t" 'intero-type-at
  "u" 'intero-uses-at
  "v" 'haskell-intero/insert-type
  "w" 'intero-cd
  "x" 'intero-repl-clear-buffer
  "y" 'hayoo
  "z" 'haskell-cabal-visit-file)

(define-key intero-repl-mode-map (kbd "C-k") 'tmux-nav-up)
(define-key intero-repl-mode-map (kbd "C-j") 'tmux-nav-down)
