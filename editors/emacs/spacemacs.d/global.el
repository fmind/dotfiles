                                        ; CONFIG

;; minors
(global-company-mode)
(global-hl-line-mode -1)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)

;; projectile
(setq projectile-globally-ignored-directories '("out"))
(setq projectile-globally-ignored-file-suffixes '("jpg" "png" "gif" "pyc"))

;; abbreviations
(setq-default abbrev-mode t)
(setq save-abbrevs 'silently)
(setq abbrev-file-name (concat SPMDIR "abbreviations"))

;; initialization
(setq vc-follow-symlinks t)

                                        ; FUNCTIONS

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

(defun my-split-and-go-to-test ()
  (interactive)
  (split-window-below-and-focus)
  (projectile-toggle-between-implementation-and-test))

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

(defun my-paste-from-clipboard ()
  "Paste from x-clipboard."
  (interactive)
  (if (display-graphic-p)
      (clipboard-yank)
      (insert (shell-command-to-string "xsel -o -b"))))

                                        ; HOOKS

;; (add-hook 'yas-after-exit-snippet-hook 'yas-reload-all)

(add-hook 'focus-out-hook (lambda () (save-some-buffers t)))

                                        ; BINDINGS

;; MOTIONS
(spacemacs/set-leader-keys "]" 'evil-avy-goto-char-2)
(spacemacs/set-leader-keys "[" 'evil-avy-goto-word-or-subword-1)
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; EDITIONS
(define-key evil-motion-state-map "U" 'redo)
(define-key evil-visual-state-map "U" 'redo)
(spacemacs/set-leader-keys "oj" 'evil-join)
(spacemacs/set-leader-keys "os" 'spacemacs/sort-lines)
(define-key evil-motion-state-map "L" 'lisp-state-toggle-lisp-state)
(define-key evil-motion-state-map "zj" 'spacemacs/evil-insert-line-below)
(define-key evil-motion-state-map "zk" 'spacemacs/evil-insert-line-above)

;; BUFFERS
(global-set-key [remap next-buffer] 'my-next-buffer)
(define-key evil-normal-state-map "K" 'my-next-buffer)
(define-key evil-normal-state-map "J" 'my-previous-buffer)
(global-set-key [remap previous-buffer] 'my-previous-buffer)

;; WINDOWS
(spacemacs/set-leader-keys "`" 'winum-select-window-0)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(spacemacs/set-leader-keys "pq" 'my-split-and-go-to-test)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(define-key evil-hybrid-state-map (kbd "C-k") 'tmux-nav-up)
(define-key evil-hybrid-state-map (kbd "C-j") 'tmux-nav-down)
(define-key evil-hybrid-state-map (kbd "C-h") 'tmux-nav-left)
(define-key evil-hybrid-state-map (kbd "C-l") 'tmux-nav-right)
(define-key evil-insert-state-map (kbd "C-k") 'tmux-nav-up)
(define-key evil-insert-state-map (kbd "C-j") 'tmux-nav-down)
(define-key evil-insert-state-map (kbd "C-h") 'tmux-nav-left)
(define-key evil-insert-state-map (kbd "C-l") 'tmux-nav-right)
(spacemacs/set-leader-keys "o0" 'eyebrowse-switch-to-window-config-0)
(spacemacs/set-leader-keys "o1" 'eyebrowse-switch-to-window-config-1)
(spacemacs/set-leader-keys "o2" 'eyebrowse-switch-to-window-config-2)
(spacemacs/set-leader-keys "o3" 'eyebrowse-switch-to-window-config-3)
(spacemacs/set-leader-keys "o4" 'eyebrowse-switch-to-window-config-4)
(spacemacs/set-leader-keys "o5" 'eyebrowse-switch-to-window-config-5)
(spacemacs/set-leader-keys "o6" 'eyebrowse-switch-to-window-config-6)
(spacemacs/set-leader-keys "L" 'spacemacs/workspaces-transient-state/body)
(spacemacs/set-leader-keys "-" 'spacemacs/workspaces-transient-state/body)

;; LISP
(spacemacs/set-leader-keys "kn" 'evil-lisp-state-sp-backward-up-sexp)

;; SHELLS
(spacemacs/set-leader-keys "\"" 'spacemacs/shell-pop-term)

;; ZOOMING
(define-key global-map (kbd "C-+") 'text-scale-increase)
(define-key global-map (kbd "C--") 'text-scale-decrease)

;; COPY/PASTE
(spacemacs/set-leader-keys "oy" 'my-copy-to-clipboard)
(spacemacs/set-leader-keys "op" 'my-paste-from-clipboard)

;; YASNIPPET
(spacemacs/set-leader-keys "ic" 'aya-create)
(spacemacs/set-leader-keys "ie" 'aya-expand)
(spacemacs/set-leader-keys "iw" 'aya-persist-snippet)
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "oi" 'my-snippet-open)

;; ABBREVIATIONS
(spacemacs/set-leader-keys "oa" 'add-mode-abbrev)
(spacemacs/set-leader-keys "oA" 'add-global-abbrev)

;; CONFIGURATIONS
(spacemacs/set-leader-keys "oc" 'my-config-open)
