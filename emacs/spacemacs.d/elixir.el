                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'elixir-mode
  "hh" 'alchemist-help
  "hi" 'alchemist-info-datatype-at-point
  "to" 'alchemist-test-toggle-test-report-display
  "\"" 'alchemist-iex-project-run
  "b" 'alchemist-iex-reload-module
  "," 'alchemist-iex-compile-this-buffer
  "l" 'alchemist-iex-send-current-line
  "L" 'alchemist-iex-send-current-line-and-go
  "r" 'alchemist-iex-send-region
  "R" 'alchemist-iex-send-region-and-go
  )
