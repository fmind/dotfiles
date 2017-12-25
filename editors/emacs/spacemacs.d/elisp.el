                                        ; CONF

(remove-hook 'emacs-lisp-mode-hook 'auto-compile-mode)

                                        ; KEYS

(spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
  "b" 'eval-buffer
  "f" 'eval-defun
  "l" 'lisp-state-eval-sexp-end-of-line
  "w" 'eval-region ;; same as clojure
  "u" 'spacemacs/ert-run-tests-buffer
  "y" 'spacemacs/eval-current-form
  "," 'eval-last-sexp)
