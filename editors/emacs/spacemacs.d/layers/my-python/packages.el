(defconst my-python-packages '(python))

(defun my-python/post-init-python ()
  (use-package python
    :defer t
    :init
    (progn
      (spacemacs/set-leader-keys-for-major-mode 'python-mode
        "," 'python-shell-send-buffer
        "-" 'spacemacs/python-remove-unused-imports
        "=" 'my-python-kill-yapify-buffers
        "B" 'python-shell-send-buffer-switch
        "C" 'spacemacs/python-execute-file
        "D" 'spacemacs/python-toggle-breakpoint
        "F" 'python-shell-send-defun-switch
        "J" 'spacemacs/jump-to-definition-other-window
        "W" 'python-shell-send-region-switch
        "Y" 'spacemacs/python-test-pdb-module
        "Y" 'spacemacs/python-test-pdb-one
        "\"" 'my-python-start-repl-with-venv
        "`" 'my-python-swith-to-repl
        "b" 'python-shell-send-buffer
        "f" 'python-shell-send-defun
        "j" 'spacemacs/jump-to-definition
        "k" 'anaconda-mode-find-assignments
        "k" 'anaconda-mode-go-back
        "l" 'anaconda-mode-find-references
        "o" 'anaconda-mode-show-doc
        "u" 'spacemacs/python-test-one
        "w" 'python-shell-send-region
        "x" 'python-toggle-breakpoint
        "y" 'spacemacs/python-test-module
        "~" 'my-python-swith-to-report
))))
