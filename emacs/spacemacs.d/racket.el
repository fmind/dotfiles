                                        ; HOOKS

;;(add-hook 'racket-mode-hook #'racket-unicode-input-method-enable)
;;(add-hook 'racket-repl-mode-hook #'racket-unicode-input-method-enable)
(add-hook 'racket-describe-mode 'disable-evil-mode)

                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'racket-mode
  "," 'racket-run
  "m" 'racket-run-and-switch-to-repl
  "d" 'spacemacs/jump-to-definition
)
