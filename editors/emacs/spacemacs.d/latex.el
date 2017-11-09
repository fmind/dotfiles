                                        ; HOOK

;; docview
(add-hook 'doc-view-mode-hook 'auto-revert-mode)

;; latex
(add-hook 'LaTeX-mode-hook 'spacemacs/toggle-visual-line-navigation-on)
