                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'haskell-mode
  "'" 'haskell-interactive-bring
  "`" 'haskell-interactive-switch
  "," 'haskell-process-load-file
  "C" 'haskell-interactive-mode-clear)

(spacemacs/set-leader-keys-for-major-mode 'haskell-interactive-mode
  "`" 'haskell-interactive-switch-back)
