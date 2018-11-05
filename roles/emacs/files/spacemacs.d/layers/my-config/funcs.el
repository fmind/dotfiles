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

(defun xclip-cut-function (text &optional push)
  (with-temp-buffer
    (insert text)
    (call-process-region (point-min) (point-max) "xclip" nil 0 nil "-i" "-selection" "clipboard")))

(defun xclip-paste-function()
  (let ((xclip-output (shell-command-to-string "xclip -o -selection clipboard")))
  (unless (string= (car kill-ring) xclip-output) xclip-output)))

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
