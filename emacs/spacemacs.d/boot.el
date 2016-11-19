                                        ; FUNCTIONS

(defun my-config-path (file)
  (expand-file-name file MYSPACE))

(defun my-config-load (file)
  (load-file (my-config-path file)))

(defun my-config-open ()
  (interactive)
  (let* ((helm-name (concat "elisp files in: " MYSPACE))
         (file (helm :sources (helm-build-sync-source helm-name
                                :fuzzy-match t
                                :candidates (lambda () (directory-files MYSPACE nil ".*el")))
                     :buffer "Helm: open configuration file")))
    (if file (find-file (my-config-path file)))))
