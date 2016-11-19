                                        ; HOOKS

(add-hook 'pandoc-mode-hook 'pandoc-load-default-settings)
(add-hook 'pandoc-mode-hook (lambda () (add-hook 'after-save-hook
                                                 (lambda () (pandoc-run-pandoc nil)) nil 'make-it-local)))

                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-minor-mode 'pandoc-mode
  "Pr" 'pandoc-run-pandoc
  "Pp" 'pandoc-convert-to-pdf)
