(defun my-elixir-switch-to-repl ()
  "Switch to elixir repl."
  (interactive)
  (switch-to-buffer "*Alchemist-IEx*"))

(defun my-elixir-switch-to-report ()
  "Switch to elixir test report."
  (interactive)
  (switch-to-buffer "*alchemist test report*"))

(defun my-elixir-open-mix-file ()
  "Open the mix.exs file of the project."
  (interactive)
  (find-file (expand-file-name "mix.exs" (projectile-project-root))))
