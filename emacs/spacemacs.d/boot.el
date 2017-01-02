                                        ; FUNCTIONS

(defun my-config-path (file)
  (expand-file-name file MYSPACE))

(defun my-config-load (file)
  (load-file (my-config-path file)))
