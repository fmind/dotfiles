                                        ; FUNCTIONS

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

                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'elixir-mode
  ;; project
  "pm" 'my-elixir-open-mix-file
  ;; tests
  "to" 'alchemist-test-toggle-test-report-display
  ;; help
  "hh" 'alchemist-help
  "hi" 'alchemist-info-datatype-at-point
  ;; goto
  "`" 'my-elixir-switch-to-repl
  "~" 'my-elixir-switch-to-report
  ;; mix
  "M" 'alchemist-mix
  ;; iex
  "\"" 'alchemist-iex-project-run
  "b" 'alchemist-iex-reload-module
  "k" 'alchemist-iex-compile-this-buffer
  "l" 'alchemist-iex-send-current-line
  "L" 'alchemist-iex-send-current-line-and-go
  "r" 'alchemist-iex-send-region
  "R" 'alchemist-iex-send-region-and-go
  )
