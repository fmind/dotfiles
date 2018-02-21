(defconst my-latex-packages '(auctex))

(defun my-latex/post-init-auctex ()
  (use-package tex
    :defer t
    :init
    (progn
      (add-hook 'doc-view-mode-hook 'auto-revert-mode)
      (add-hook 'LaTeX-mode-hook 'spacemacs/toggle-golden-ratio-off)
      (add-hook 'LaTeX-mode-hook 'spacemacs/toggle-visual-line-navigation-on))))
