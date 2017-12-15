                                        ; CONF

;; minors
(golden-ratio-mode)
(global-company-mode)
(global-hl-line-mode -1)
(setq create-lockfiles nil)
(ranger-override-dired-mode t)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)

;; projectile
(setq projectile-globally-ignored-directories '("out" "target"))
(setq projectile-globally-ignored-file-suffixes '("jpg" "png" "gif" "pyc"))

;; abbreviations
(setq-default abbrev-mode t)
(setq save-abbrevs 'silently)
(setq abbrev-file-name (concat SPMDIR "abbreviations"))

;; initialization
(setq vc-follow-symlinks t)

                                        ; FUNS

(defun my-config-open (file)
  "Open a configuration file in spacemacs directory."
  (interactive (list nil))
  (if file (find-file)
    (helm-find-files-1 SPMDIR)))

(defun my-snippet-open (file)
  "Open a snippet file in spacemacs directory."
  (interactive (list nil))
  (if file (find-file file)
    (let* ((snipdir (format "%s/%S/" auto-completion-private-snippets-directory major-mode)))
      (unless (file-exists-p snipdir) (make-directory snipdir t))
      (helm-find-files-1 snipdir)))
  ;; insert default content to the snippet
  (when (= (buffer-size (current-buffer)) 0)
    (insert (format "# -*- mode: snippet -*-\n# contributor: fmind\n# name: %s\n# key: %s\n# --\n"
                    (buffer-name) (buffer-name)))))

(defun my-eshell-opener (splitter)
  "Open a window with splitter and start an eshell."
  `(lambda () (interactive) (,splitter) (eshell)))

(defun my-emacs-buffer-p (buf-name)
  "Test if a buffer is associated to emacs."
  (and (string-prefix-p "*" buf-name)
       (string-suffix-p "*" buf-name)))

(defun my-user-buffer-p (buf-name)
  "Test if a buffer is associated to the user."
  (not (my-emacs-buffer-p buf-name)))

(defun my-next-buffer ()
  "next-buffer, skip emacs and dired buffers."
  (interactive)
  (let ((current (buffer-name)))
    (while (progn
             (next-buffer)
             (when (not (string= current (buffer-name)))
               (or (my-emacs-buffer-p (buffer-name))
                   (string= major-mode "dired-mode")))))))

(defun my-previous-buffer ()
  "next-buffer, skip emacs and dired buffers."
  (interactive)
  (let ((current (buffer-name)))
    (while (progn
             (previous-buffer)
             (when (not (string= current (buffer-name)))
               (or (my-emacs-buffer-p (buffer-name))
                   (string= major-mode "dired-mode")))))))

(defun my-copy-to-clipboard ()
  "Copy to x-clipboard."
  (interactive)
  (if (display-graphic-p)
      (progn
        (message "Yanked region to x-clipboard!")
        (call-interactively 'clipboard-kill-ring-save))
    (if (region-active-p)
        (progn
          (shell-command-on-region (region-beginning) (region-end) "xsel -i -b")
          (message "Yanked region to clipboard!")
          (deactivate-mark))
        (message "No region active; can't yank to clipboard!"))))

(defun my-cut-to-clipboard ()
  "Cut to x-clipboard."
  (interactive)
  (my-copy-to-clipboard)
  (delete-region (region-beginning) (region-end)))

(defun my-paste-from-clipboard ()
  "Paste from x-clipboard."
  (interactive)
  (if (display-graphic-p)
      (clipboard-yank)
      (insert (shell-command-to-string "xsel -o -b"))))

                                        ; HOOK

(add-hook 'focus-out-hook (lambda () (save-some-buffers t)))

                                        ; KEYS

;; FILES
(define-key evil-normal-state-map "-" 'deer)
(define-key evil-visual-state-map "-" 'deer)
(define-key evil-normal-state-map "_" 'ranger)
(define-key evil-visual-state-map "_" 'ranger)
(spacemacs/set-leader-keys "fa" 'fasd-find-file)
(spacemacs/set-leader-keys "fi" 'fasd-find-file-only)
(spacemacs/set-leader-keys "fd" 'fasd-find-directory-only)

;; MOTIONS
(define-key evil-normal-state-map "s" 'evil-snipe-s)
(define-key evil-normal-state-map "S" 'evil-snipe-S)
(define-key evil-normal-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-normal-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; EDITIONS
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

;; BUFFERS
(define-key evil-normal-state-map "M" 'helm-mini)
(define-key evil-visual-state-map "M" 'helm-mini)
(define-key evil-normal-state-map "L" 'my-next-buffer)
(define-key evil-visual-state-map "L" 'my-next-buffer)
(define-key evil-normal-state-map "H" 'my-previous-buffer)
(define-key evil-visual-state-map "H" 'my-previous-buffer)

;; ;; WINDOWS
(spacemacs/set-leader-keys "H" 'split-window-right)
(spacemacs/set-leader-keys "K" 'split-window-below)
(spacemacs/set-leader-keys "L" 'split-window-right-and-focus)
(spacemacs/set-leader-keys "J" 'split-window-below-and-focus)
(spacemacs/set-leader-keys "`" 'winum-select-window-0)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(define-key evil-hybrid-state-map (kbd "C-k") 'tmux-nav-up)
(define-key evil-hybrid-state-map (kbd "C-j") 'tmux-nav-down)
(define-key evil-hybrid-state-map (kbd "C-h") 'tmux-nav-left)
(define-key evil-hybrid-state-map (kbd "C-l") 'tmux-nav-right)
(spacemacs/set-leader-keys "o0" 'eyebrowse-switch-to-window-config-0)
(spacemacs/set-leader-keys "o1" 'eyebrowse-switch-to-window-config-1)
(spacemacs/set-leader-keys "o2" 'eyebrowse-switch-to-window-config-2)
(spacemacs/set-leader-keys "o3" 'eyebrowse-switch-to-window-config-3)
(spacemacs/set-leader-keys "o4" 'eyebrowse-switch-to-window-config-4)
(spacemacs/set-leader-keys "o5" 'eyebrowse-switch-to-window-config-5)
(spacemacs/set-leader-keys "o6" 'eyebrowse-switch-to-window-config-6)
(spacemacs/set-leader-keys "=" 'spacemacs/layouts-transient-state/body)
(spacemacs/set-leader-keys "-" 'spacemacs/workspaces-transient-state/body)

;; SHELLS
(spacemacs/set-leader-keys "aH" (my-eshell-opener 'split-window-right))
(spacemacs/set-leader-keys "aK" (my-eshell-opener 'split-window-below))
(spacemacs/set-leader-keys "aL" (my-eshell-opener 'split-window-right-and-focus))
(spacemacs/set-leader-keys "aJ" (my-eshell-opener 'split-window-below-and-focus))
(use-package eshell
    :defer t
    :init
  (progn
    (evil-define-key 'normal eshell-mode-map
      (kbd "C-n") 'eshell-next-matching-input-from-input
      (kbd "C-p") 'eshell-previous-matching-input-from-input)
    (evil-define-key 'hybrid
      (kbd "C-n") 'eshell-next-matching-input-from-input
      (kbd "C-p") 'eshell-previous-matching-input-from-input)))

;; SERVER
(spacemacs/set-leader-keys "qq" 'spacemacs/frame-killer)
(spacemacs/set-leader-keys "qQ" 'spacemacs/prompt-kill-emacs)
(define-key evil-normal-state-map (kbd "C-z") 'suspend-frame)
(define-key evil-visual-state-map (kbd "C-z") 'suspend-frame)
(define-key evil-hybrid-state-map (kbd "C-z") 'suspend-frame)

;; ZOOMING
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)

;; CUSTOMS
(spacemacs/set-leader-keys "oj" 'evil-join)
(spacemacs/set-leader-keys "os" 'spacemacs/sort-lines)

;; COPY/PASTE
(spacemacs/set-leader-keys "ox" 'my-cut-to-clipboard)
(spacemacs/set-leader-keys "oy" 'my-copy-to-clipboard)
(spacemacs/set-leader-keys "op" 'my-paste-from-clipboard)

;; YASNIPPET
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "oi" 'my-snippet-open)

;; ABBREVIATIONS

(spacemacs/set-leader-keys "oa" 'add-mode-abbrev)
(spacemacs/set-leader-keys "oA" 'add-global-abbrev)

;; CONFIGURATIONS
(spacemacs/set-leader-keys "oc" 'my-config-open)
