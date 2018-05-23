(defconst my-tmux-packages '(tmux))

(defun my-tmux/post-init-tmux ()
  (use-package tmux
    :defer t
    :init
    (progn
      (define-key evil-hybrid-state-map (kbd "C-k") 'tmux-nav-up)
      (define-key evil-hybrid-state-map (kbd "C-j") 'tmux-nav-down)
      (define-key evil-hybrid-state-map (kbd "C-h") 'tmux-nav-left)
      (define-key evil-hybrid-state-map (kbd "C-l") 'tmux-nav-right))))
