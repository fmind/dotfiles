                                        ; CONFIGURATIONS

(setq alchemist-test-ask-about-save nil)
(setq alchemist-hooks-compile-on-save t)

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
;;   ;; tests
;;   "to" 'alchemist-test-toggle-test-report-display
  ;; help
  "?" 'alchemist-help
  "H" 'alchemist-help-search-at-point
;;   "hi" 'alchemist-info-datatype-at-point
  ;; goto
  "S" 'my-elixir-switch-to-repl
  "T" 'my-elixir-switch-to-report
  ;; mix
  ";" 'alchemist-mix
  "C" 'alchemist-mix-compile
  "M" 'my-elixir-open-mix-file
  ;; iex
  "r" 'alchemist-iex-send-region
  "R" 'alchemist-iex-send-region-and-go
  "l" 'alchemist-iex-send-current-line
  "L" 'alchemist-iex-send-current-line-and-go
  "," 'alchemist-iex-compile-this-buffer
  "." 'alchemist-iex-compile-this-buffer-and-go
;;   "\"" 'alchemist-iex-project-run
;;   "b" 'alchemist-iex-reload-module
;;   "k" 'alchemist-iex-compile-this-buffer
;;   "l" 'alchemist-iex-send-current-line
;;   "L" 'alchemist-iex-send-current-line-and-go
;;   "r" 'alchemist-iex-send-region
;;   "R" 'alchemist-iex-send-region-and-go
  )
