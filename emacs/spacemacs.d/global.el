                                        ; CONFIG

(global-company-mode)
(setq vc-follow-symlinks t)
(setq smooth-scroll-margin 10)
(setq powerline-default-separator 'brace)
(spacemacs/toggle-evil-cleverparens-on)

                                        ; KEYBINDINGS

;; MOTIONS
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; ZOOMING
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)

;; YASNIPPET
(spacemacs/set-leader-keys "ic" 'aya-create)
(spacemacs/set-leader-keys "ie" 'aya-expand)
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "iw" 'aya-persist-snippet)
