(defconst my-python-packages '(python))

(defun my-python/post-init-python ()
  (use-package python
    :defer t
    :init
    (progn
      (spacemacs/set-leader-keys-for-major-mode 'python-mode
        "`" 'my-python-swith-to-repl
        "~" 'my-python-swith-to-report
        "\"" 'my-python-start-repl-with-venv
        "G" 'spacemacs/jump-to-definition
        "C" 'spacemacs/python-execute-file
        "D" 'spacemacs/python-toggle-breakpoint
        "I" 'spacemacs/python-remove-unused-imports
        "Y" 'my-python-kill-yapify-buffers
        "," 'python-shell-send-buffer
        "b" 'python-shell-send-buffer
        "B" 'python-shell-send-buffer-switch
        "f" 'python-shell-send-defun
        "F" 'python-shell-send-defun-switch
        "n" 'python-shell-send-region
        "N" 'python-shell-send-region-switch))))
