                                        ; BINDINGS

(spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
  "f" 'eval-defun
  "," 'eval-buffer
  "r" 'eval-region
  "i" 'eval-last-sexp
  "x" 'eval-expression
  )
