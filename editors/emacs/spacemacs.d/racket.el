                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'racket-mode
  "`" 'racket-unvisit
  "w" 'racket-doc
  "d" 'racket-describe
  "l" 'racket-insert-lambda
  "O" 'racket-visit-module
  "o" 'racket-open-require-path
  "," 'racket-run
  "b" 'racket-run
  "B" 'racket-run-and-switch-to-repl
  "e" 'racket-send-last-sexp
  "E" 'racket-send-last-sexp-focus
  "f" 'racket-send-definition
  "F" 'racket-send-definition-focus
  "r" 'racket-send-region
  "R" 'racket-send-region-focus
  "a" 'racket-test
  "A" 'racket-test-with-coverage
  )
