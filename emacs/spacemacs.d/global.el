                                        ; CONFIG

;; global
(global-company-mode)
(setq powerline-default-separator 'brace)
(spacemacs/toggle-evil-cleverparens-on)

;; shell initilization
(setq vc-follow-symlinks t)
;;(exec-path-from-shell-initialize)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)

                                        ; FUNCTIONS

(defun my-config-open ()
  (interactive)
  (let* ((helm-name (concat "elisp files in: " MYSPACE))
         (file (helm :sources (helm-build-sync-source helm-name
                                :fuzzy-match t
                                :candidates (lambda () (directory-files MYSPACE nil ".*el")))
                     :buffer "Helm: open configuration file")))
    (if file (find-file (my-config-path file)))))


                                        ; KEYBINDINGS

;; MOTIONS
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; CONFS
(spacemacs/set-leader-keys "oc" 'my-config-open)

;; JUMPS
(spacemacs/set-leader-keys "[" 'evil-avy-goto-word-or-subword-1)
(spacemacs/set-leader-keys "]" 'evil-avy-goto-char)

;; SHELLS
(spacemacs/set-leader-keys "\"" 'spacemacs/shell-pop-term)

;; ZOOMING
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)

;; YASNIPPET
(spacemacs/set-leader-keys "ic" 'aya-create)
(spacemacs/set-leader-keys "ie" 'aya-expand)
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "iw" 'aya-persist-snippet)
