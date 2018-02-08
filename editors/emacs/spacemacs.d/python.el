                                        ; FUNS

(defun my-python-swith-to-repl ()
  (interactive)
  (switch-to-buffer "*Python*"))

(defun my-python-kill-yapify-buffers ()
  (interactive)
  (cf-flet ((kill-buffer-ask (buffer) (kill-buffer buffer)))
    (kill-matching-buffers  ".*\*yapfify\*.*")))

(defun my-python-start-repl-with-venv ()
  (interactive)
  (pyvenv-activate (concat (projectile-project-root) "venv"))
  (python-start-or-switch-repl))

                                        ; KEYS

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
  "N" 'python-shell-send-region-switch)

                                        ; NOTE

;; (setq ein:use-auto-complete t)
;; (spacemacs/set-leader-keys "oN" 'ein:notebooklist-open)
