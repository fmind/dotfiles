(defconst my-evil-snipe-packages '(evil-snipe))

(defun my-evil-snipe/post-init-evil-snipe ()
  (use-package evil-snipe
    :defer t
    :init
    (progn
      (define-key evil-normal-state-map "s" 'evil-snipe-s)
      (define-key evil-normal-state-map "S" 'evil-snipe-S))))
