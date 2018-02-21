(defun my-emacs-buffer-p (buf-name)
  "Test if a buffer is an emacs buffer."
  (and (string-prefix-p "*" buf-name)
       (string-suffix-p "*" buf-name)))

(defun my-user-buffer-p (buf-name)
  "Test if a buffer is an user buffer."
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

(defun my-copy-to-clipboard ()
  "Copy to X-clipboard."
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

(defun my-config-open (file)
  "Open a configuration file in spacemacs directory."
  (interactive (list nil))
  (if file
    (find-file file)
    (helm-find-files-1 (concat PRIVATE "/layers/"))))

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
