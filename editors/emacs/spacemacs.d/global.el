                                        ; CONFIG

;; minors
(global-company-mode)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)

;; abbreviations
(setq-default abbrev-mode t)
(setq save-abbrevs 'silently)
(setq abbrev-file-name (concat SPMDIR "abbreviations"))

;; initialization
(setq vc-follow-symlinks t)
(setq-default dotspacemacs-persistent-server t)

                                        ; FUNCTIONS

(defun my-config-open ()
  "Open a configuration file in spacemacs directory."
  (interactive)
  (let* ((helm-name (concat "elisp files in: " SPMDIR))
         (file (helm :sources (helm-build-sync-source helm-name
                                :fuzzy-match t
                                :candidates (lambda () (directory-files SPMDIR nil ".*el")))
                     :buffer "Helm: open configuration file")))
    (if file (find-file (my-config-path file)))))

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

(add-hook 'focus-out-hook (lambda () (save-some-buffers t)))

                                        ; BINDINGS

;; MOTIONS
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; EDITIONS
(spacemacs/set-leader-keys "oj" 'evil-join)

;; BUFFERS
(global-set-key [remap next-buffer] 'my-next-buffer)
(define-key evil-normal-state-map "K" 'my-next-buffer)
(global-set-key [remap previous-buffer] 'my-previous-buffer)
(define-key evil-normal-state-map "J" 'my-previous-buffer)

;; WINDOWS
(spacemacs/set-leader-keys "`" 'winum-select-window-0)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(spacemacs/set-leader-keys "pq" 'my-split-and-go-to-test)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(spacemacs/set-leader-keys "o0" 'eyebrowse-switch-to-window-config-0)
(spacemacs/set-leader-keys "o1" 'eyebrowse-switch-to-window-config-1)
(spacemacs/set-leader-keys "o2" 'eyebrowse-switch-to-window-config-2)
(spacemacs/set-leader-keys "o3" 'eyebrowse-switch-to-window-config-3)
(spacemacs/set-leader-keys "o4" 'eyebrowse-switch-to-window-config-4)
(spacemacs/set-leader-keys "o5" 'eyebrowse-switch-to-window-config-5)
(spacemacs/set-leader-keys "o6" 'eyebrowse-switch-to-window-config-6)
(spacemacs/set-leader-keys "L" 'spacemacs/workspaces-transient-state/body)

;; LISP
(spacemacs/set-leader-keys "kn" 'evil-lisp-state-sp-backward-up-sexp)

;; JUMPS
(spacemacs/set-leader-keys "]" 'evil-avy-goto-char)
(spacemacs/set-leader-keys "[" 'evil-avy-goto-word-or-subword-1)

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
(spacemacs/set-leader-keys "ir" 'yas-reload-all)
(spacemacs/set-leader-keys "iw" 'aya-persist-snippet)

;; ABBREVIATIONS
(spacemacs/set-leader-keys "oa" 'add-mode-abbrev)
(spacemacs/set-leader-keys "oA" 'add-global-abbrev)

;; CONFIGURATIONS
(spacemacs/set-leader-keys "oc" 'my-config-open)
