(defconst my-pandoc-packages '(pandoc-mode))

(defun my-pandoc/post-init-pandoc-mode ()
  (use-package pandoc-mode
    :defer t
    :init
    (progn
      (add-hook 'pandoc-mode-hook 'pandoc-load-default-settings)
      (add-hook 'pandoc-mode-hook (lambda () (add-hook 'after-save-hook
                                                      (lambda () (pandoc-run-pandoc nil)) nil 'make-it-local)))
      (spacemacs/set-leader-keys-for-minor-mode 'pandoc-mode
        "Pr" 'pandoc-run-pandoc
        "Pp" 'pandoc-convert-to-pdf))))
