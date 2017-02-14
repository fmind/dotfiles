                                        ; CONFIGS

(setq cider-repl-display-help-banner nil)

                                        ; FUNCTIONS

(defun my-init-cider ()
  (interactive)
  (cider-switch-to-repl-buffer)
  (split-window-below-and-focus)
  (switch-to-buffer "*cider-test-report*")
  (select-window-1)
  (split-window-below-and-focus)
  (projectile-toggle-between-implementation-and-test)
  (select-window-1))

                                        ; HOOKS

(dolist (mode '(clojure-mode clojurec-mode clojurescript-mode clojurex-mode cider-repl-mode))
  (add-hook mode #'golden-ratio-mode)
  (add-hook mode #'aggressive-indent-mode)
  (add-hook mode #'evil-cleverparens-mode))

(add-to-list 'magic-mode-alist '(".* boot" . clojure-mode))   ; boot scripts
(add-to-list 'auto-mode-alist '("\\.boot\\'" . clojure-mode)) ; boot files

                                        ; KEYBINDINGS

(dolist (mode '(clojure-mode clojurec-mode clojurescript-mode clojurex-mode cider-repl-mode))
  (spacemacs/set-leader-keys-for-major-mode mode
    "," 'cider-load-buffer
    "@" 'my-init-cider
    "v" 'cider-find-var
    "i" 'cider-inspect
    "b" 'cider-pop-back
    "n" 'cider-repl-set-ns
    "N" 'cider-load-buffer-and-repl-set-ns
    "c" 'cider-repl-clear-buffer
    "u" 'cider-switch-to-repl-buffer
    "te" 'cider-test-run-tests
    "dv" 'cider-toggle-trace-var
    "dn" 'cider-toggle-trace-ns
    "tn" 'cider-test-run-ns-tests
    "tp" 'cider-test-run-project-tests
    "tl" 'cider-test-run-loaded-tests))
