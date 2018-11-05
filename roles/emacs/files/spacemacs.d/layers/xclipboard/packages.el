(setq xclipboard-packages '(xclip))

(defun xclipboard/init-xclip ()
  (use-package xclip
    :config (xclip-mode 1)))
