                                        ; CONFIGS

(setq cider-repl-use-pretty-printing t)
(setq cider-repl-display-help-banner nil)
(setq cider-repl-display-in-current-window t)

                                        ; FUNCTIONS

;; ;; TODO: replace by layout
;; (defun my-init-cider ()
;;   (interactive)
;;   (cider-switch-to-repl-buffer)
;;   (split-window-below-and-focus)
;;   (switch-to-buffer "*cider-test-report*")
;;   (select-window-1)
;;   (split-window-below-and-focus)
;;   (projectile-toggle-between-implementation-and-test)
;;   (select-window-1))

(defun my-load-buffer-and-repl-set-ns ()
  (interactive)
  (cider-load-buffer)
  (cider-repl-set-ns (cider-current-ns)))

(defun my-eval-and-next ()
  (interactive)
  (cider-eval-last-sexp)
  (evil-lisp-state-sp-forward-sexp)
  (evil-lisp-state/quit)
  (evil-next-visual-line))

(defun my-eval-and-previous()
  (interactive)
  (cider-eval-last-sexp)
  (evil-lisp-state-sp-backward-sexp)
  (evil-lisp-state/quit)
  (evil-previous-visual-line))

(defun my-defonce-toggle ()
  (interactive)
  (let* ((line (thing-at-point 'line t))
         (line (cond
                ((string-match-p "^(def " line) (replace-regexp-in-string "(def " "(defonce " line))
                ((string-match-p "^(defonce " line) (replace-regexp-in-string "(defonce " "(def " line))
                (t line))))
    (kill-whole-line) (insert line) (previous-line)))

                                        ; HOOKS

(dolist (mode '(clojure-mode clojurec-mode clojurescript-mode clojurex-mode cider-repl-mode))
  ;; (add-hook mode #'golden-ratio-mode)
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
    "D" 'my-defonce-toggle
    "j" 'my-eval-and-next
    "k" 'my-eval-and-previous
    "N" 'my-load-buffer-and-repl-set-ns
    "c" 'cider-repl-clear-buffer
    "u" 'cider-switch-to-repl-buffer
    "R" 'cider-test-show-report
    "te" 'cider-test-run-tests
    "dv" 'cider-toggle-trace-var
    "dn" 'cider-toggle-trace-ns
    "tn" 'cider-test-run-ns-tests
    "tp" 'cider-test-run-project-tests
    "tl" 'cider-test-run-loaded-tests))

;; specific to cider-repl-mode-map

(defun my-cider-mode-hook ()
    (define-key cider-repl-mode-map (kbd "C-p") #'cider-repl-backward-input)
    (define-key cider-repl-mode-map (kbd "C-p") #'cider-repl-backward-input)
  )

(add-hook 'cider-repl-mode-hook 'my-cider-mode-hook)
