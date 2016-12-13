                                        ; CONFIG

(global-company-mode)
(setq vc-follow-symlinks t)
(setq powerline-default-separator 'brace)
(spacemacs/toggle-evil-cleverparens-on)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)


                                        ; KEYBINDINGS

;; MOTIONS
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; JUMPS
(spacemacs/set-leader-keys "[" 'evil-avy-goto-word-or-subword-1)
(spacemacs/set-leader-keys "]" 'evil-avy-goto-char)

;; ZOOMING
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)

;; YASNIPPET
(spacemacs/set-leader-keys "ic" 'aya-create)
(spacemacs/set-leader-keys "ie" 'aya-expand)
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "iw" 'aya-persist-snippet)
