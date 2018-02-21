;; CUSTOMS {{{
(spacemacs/set-leader-keys "oj" 'evil-join)
(spacemacs/set-leader-keys "oc" 'my-config-open)
(spacemacs/set-leader-keys "os" 'spacemacs/sort-lines)
;;}}}
;; ZOOMING {{{
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)
;;}}}
;; BUFFERS {{{
(define-key evil-normal-state-map "M" 'helm-mini)
(define-key evil-visual-state-map "M" 'helm-mini)
(define-key evil-normal-state-map "L" 'my-next-buffer)
(define-key evil-visual-state-map "L" 'my-next-buffer)
(define-key evil-normal-state-map "H" 'my-previous-buffer)
(define-key evil-visual-state-map "H" 'my-previous-buffer)
;;}}}
;; SERVERS {{{
(define-key evil-normal-state-map (kbd "C-z") 'suspend-frame)
(define-key evil-visual-state-map (kbd "C-z") 'suspend-frame)
(define-key evil-hybrid-state-map (kbd "C-z") 'suspend-frame)
;;}}}
;; WINDOWS {{{
(spacemacs/set-leader-keys "H" 'split-window-right)
(spacemacs/set-leader-keys "K" 'split-window-below)
(spacemacs/set-leader-keys "L" 'split-window-right-and-focus)
(spacemacs/set-leader-keys "J" 'split-window-below-and-focus)
(spacemacs/set-leader-keys "`" 'winum-select-window-0)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(spacemacs/set-leader-keys "o0" 'eyebrowse-switch-to-window-config-0)
(spacemacs/set-leader-keys "o1" 'eyebrowse-switch-to-window-config-1)
(spacemacs/set-leader-keys "o2" 'eyebrowse-switch-to-window-config-2)
(spacemacs/set-leader-keys "o3" 'eyebrowse-switch-to-window-config-3)
(spacemacs/set-leader-keys "o4" 'eyebrowse-switch-to-window-config-4)
(spacemacs/set-leader-keys "o5" 'eyebrowse-switch-to-window-config-5)
(spacemacs/set-leader-keys "o6" 'eyebrowse-switch-to-window-config-6)
(spacemacs/set-leader-keys "=" 'spacemacs/layouts-transient-state/body)
(spacemacs/set-leader-keys "-" 'spacemacs/workspaces-transient-state/body)
;;}}}
;; MOTIONS {{{
(define-key evil-normal-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-normal-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)
;;}}}
;; EDITIONS {{{
(define-key evil-normal-state-map "U" 'redo)
(define-key evil-visual-state-map "U" 'redo)
(define-key evil-normal-state-map "K" 'evil-scroll-up)
(define-key evil-visual-state-map "K" 'evil-scroll-up)
(define-key evil-normal-state-map "J" 'evil-scroll-down)
(define-key evil-visual-state-map "J" 'evil-scroll-down)
(define-key evil-normal-state-map "Q" 'lisp-state-toggle-lisp-state)
(define-key evil-visual-state-map "Q" 'lisp-state-toggle-lisp-state)
(define-key evil-normal-state-map "zj" 'spacemacs/evil-insert-line-below)
(define-key evil-normal-state-map "zk" 'spacemacs/evil-insert-line-above)
;;}}}
;; SNIPPETS {{{
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "oi" 'my-snippet-open)
;;}}}
;; COPY/PASTE {{{
(spacemacs/set-leader-keys "ox" 'my-cut-to-clipboard)
(spacemacs/set-leader-keys "oy" 'my-copy-to-clipboard)
(spacemacs/set-leader-keys "op" 'my-paste-from-clipboard)
;;}}}
;; ABBREVIATIONS {{{
(spacemacs/set-leader-keys "oa" 'add-mode-abbrev)
(spacemacs/set-leader-keys "oA" 'add-global-abbrev)
;;}}}
