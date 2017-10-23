                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
  "b" 'eval-buffer
  "f" 'eval-defun
  "l" 'lisp-state-eval-sexp-end-of-line
  "r" 'eval-region
  "u" 'spacemacs/ert-run-tests-buffer
  "y" 'spacemacs/eval-current-form
  "," 'eval-last-sexp
  )
