                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'haskell-mode
  "-" 'haskell-mode-jump-to-filename-in-string
  "=" 'haskell-mode-stylish-buffer
  "\"" 'haskell-intero/display-repl
  "'" 'haskell-intero/pop-to-repl
  "," 'intero-repl-load
  "A" 'haskell-process-cabal
  "B" 'haskell-process-cabal-build
  "C" 'intero-repl-clear-buffer
  ;; "D"
  ;; "E"
  ;; "F" RESERVED -- stylish
  ;; "G"
  ;; "H"
  ;; "I"
  ;; "J"
  "K" 'intero-destroy
  "L" 'intero-list-buffers
  ;; "M"
  ;; "N"
  ;; "O"
  ;; "P"
  ;; "Q"
  "R" 'intero-restart
  ;; "S"
  ;; "T"
  ;; "U"
  "V" 'haskell-cabal-visit-file
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
  "l" 'hlint-refactor-refactor-at-point
  "m" 'intero-expand-splice-at-point
  "n" 'haskell-navigate-imports
  "o" 'intero-info
  "p" 'helm-hoogle
  "q" 'intero-targets
  ;; "r" RESERVED -- refactor
  ;; "s" RESERVED -- repl
  "t" 'intero-type-at
  "u" 'intero-uses-at
  "v" 'haskell-intero/insert-type
  "w" 'intero-cd
  "x" 'haskell-compile
  "y" 'hayoo
  "z" 'intero-devel-reload
  )

(spacemacs/set-leader-keys-for-major-mode 'haskell-interactive-mode
  "," 'haskell-interactive-switch-back)
