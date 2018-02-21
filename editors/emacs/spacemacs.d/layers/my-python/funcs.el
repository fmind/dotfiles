(defun my-python-swith-to-repl ()
  (interactive)
  (switch-to-buffer "*Python*"))

(defun my-python-kill-yapify-buffers ()
  (interactive)
  (kill-matching-buffers  ".*\*yapfify\*.*"))

(defun my-python-start-repl-with-venv ()
  (interactive)
  (pyvenv-activate (concat (projectile-project-root) "venv"))
  (python-start-or-switch-repl))
