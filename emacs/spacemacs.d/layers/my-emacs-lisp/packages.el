(defconst my-emacs-lisp-packages '(emacs-lisp))

(defun my-emacs-lisp/post-init-emacs-lisp ()
  (use-package emacs-lisp
    :defer t
    :init
    (progn
      (remove-hook 'emacs-lisp-mode-hook 'auto-compile-mode)
      (spacemacs/set-leader-keys-for-major-mode 'emacs-lisp-mode
        "," 'eval-last-sexp
        "b" 'eval-buffer
        "f" 'eval-defun
        "l" 'lisp-state-eval-sexp-end-of-line
        "u" 'spacemacs/ert-run-tests-buffer
        "w" 'eval-region ;; same as clojure
        "y" 'spacemacs/eval-current-form))))
