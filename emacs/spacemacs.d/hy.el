                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'hy-mode
  "'" 'inferior-lisp
  "," 'lisp-load-file
  "B" 'switch-to-lisp
  "e" 'lisp-eval-last-sexp
  "f" 'lisp-eval-defun
  "F" 'lisp-eval-defun-and-go
  "r" 'lisp-eval-region
  "R" 'lisp-eval-defun-and-go)
