(setq clipboard-packages '(xclip))

(defun clipboard/init-xclip ()
  (use-package xclip
    :config (xclip-mode 1)))
