                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
  "!" 'ielm
  "," 'spacemacs/eval-current-form-sp
  "f" 'eval-defun
  "b" 'eval-buffer
  "r" 'eval-region
  "l" 'eval-last-sexp
  ";" 'eval-expression
  )
