                                        ; CONFIG

;; global modes
(global-company-mode)
(setq powerline-default-separator 'brace)
(spacemacs/toggle-evil-cleverparens-on)

;; scrolling
(setq scroll-margin 10)
(setq scroll-conservatively 10)


;; shell initilization
(setq vc-follow-symlinks t)

;; abbreviations
(setq-default abbrev-mode t)
(setq save-abbrevs 'silently)
(setq abbrev-file-name (concat MYSPACE "abbreviations"))

                                        ; FUNCTIONS

(defun my-config-open ()
  "Open a configuration file in MYSPACE with helm."
  (interactive)
  (let* ((helm-name (concat "elisp files in: " MYSPACE))
         (file (helm :sources (helm-build-sync-source helm-name
                                :fuzzy-match t
                                :candidates (lambda () (directory-files MYSPACE nil ".*el")))
                     :buffer "Helm: open configuration file")))
    (if file (find-file (my-config-path file)))))

(defun my-emacs-buffer-p (buf-name)
  "Test if a buffer is from emacs given its name."
  (and (string-prefix-p "*" buf-name)
       (string-suffix-p "*" buf-name)))

(defun my-user-buffer-p (buf-name)
  "Test if a buffer is from the user given its name."
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

                                        ; HOOKS

(add-hook 'focus-out-hook (lambda () (save-some-buffers t)))

                                        ; KEYBINDINGS

;; MOTIONS
(define-key evil-motion-state-map "j" 'evil-next-visual-line)
(define-key evil-visual-state-map "j" 'evil-next-visual-line)
(define-key evil-motion-state-map "k" 'evil-previous-visual-line)
(define-key evil-visual-state-map "k" 'evil-previous-visual-line)

;; BUFFERS
(global-set-key [remap next-buffer] 'my-next-buffer)
(global-set-key [remap previous-buffer] 'my-previous-buffer)

;; WINDOWS
(spacemacs/set-leader-keys "`" 'select-window-0)
(spacemacs/set-leader-keys "wq" 'kill-buffer-and-window)
(spacemacs/set-leader-keys "pq" 'my-split-and-go-to-test)

;; LISP
(spacemacs/set-leader-keys "kn" 'evil-lisp-state-sp-backward-up-sexp)

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

;; ABBREVIATIONS
(spacemacs/set-leader-keys "oa" 'add-mode-abbrev)
(spacemacs/set-leader-keys "oA" 'add-global-abbrev)

;; CONFIGURATIONS
(spacemacs/set-leader-keys "oc" 'my-config-open)
