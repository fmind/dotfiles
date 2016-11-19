                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
  "," 'spacemacs/eval-current-form-sp
  "!" 'ielm
  "r" 'eval-region
  "b" 'eval-buffer
  "f" 'eval-defun
  ";" 'eval-expression
  "l" 'eval-last-sexp)
