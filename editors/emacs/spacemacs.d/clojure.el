                                        ; CONF

(setq cider-save-file-on-load t)
(setq cider-prompt-for-symbol nil)
(setq nrepl-hide-special-buffers t)
(setq cider-repl-use-pretty-printing t)
(setq cider-repl-display-help-banner nil)
(setq cider-repl-display-in-current-window t)
(setq cider-cljs-lein-repl "(do (use 'figwheel-sidecar.repl-api) (start-figwheel!) (cljs-repl))")

                                        ; FUNS

(defun my-cider-connect ()
  (interactive)
  (let ((port (f-read-text ".nrepl-port")))
    (cider-connect "localhost" port ".")))

(defun my-defonce-toggle ()
  (interactive)
  (let* ((line (thing-at-point 'line t))
         (line (cond
                ((string-match-p "^(def " line) (replace-regexp-in-string "(def " "(defonce " line))
                ((string-match-p "^(defonce " line) (replace-regexp-in-string "(defonce " "(def " line))
                (t line))))
    (kill-whole-line) (insert line) (previous-line)))

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

(defun my-load-buffer-and-repl-set-ns ()
  (interactive)
  (cider-load-buffer)
  (cider-repl-set-ns (cider-current-ns)))

                                        ; HOOK

(dolist (mode '(clojure-mode clojurec-mode clojurescript-mode clojurex-mode cider-repl-mode))
  (add-hook mode #'aggressive-indent-mode)
  (add-hook mode #'evil-cleverparens-mode))

(defun my-cider-mode-hook ()
    (define-key cider-repl-mode-map (kbd "C-n") #'cider-repl-forward-input)
    (define-key cider-repl-mode-map (kbd "C-p") #'cider-repl-backward-input))

(add-hook 'cider-repl-mode-hook 'my-cider-mode-hook)

                                        ; KEYS

(dolist (mode '(clojure-mode clojurec-mode clojurescript-mode clojurex-mode cider-repl-mode))
  (spacemacs/set-leader-keys-for-major-mode mode
    "`" 'cider-pop-back
    "," 'cider-load-buffer
    "!" 'my-cider-connect
    "A" 'clojure-align
    ;; "B"
    "C" 'cider-repl-clear-buffer
    "D" 'my-defonce-toggle
    ;; "E"
    "F" 'cider-find-var
    "G" 'cider-grimoire
    "H" 'cider-doc
    "I" 'cider-inspect
    "J" 'cider-javadoc
    "K" 'cider-test-show-report
    "L" 'cider-enlighten-mode
    "M" 'cider-macroexpand-all
    "N" 'cider-browse-ns
    "O" 'cljr-hotload-dependency
    "P" 'spacemacs/cider-toggle-repl-pretty-printing
    "Q" 'cider-quit
    ;; "R"
    "S" 'cider-switch-to-repl-buffer
    ;; "T" RESERVED
    "U" 'cider-auto-test-mode
    "V" 'cider-toggle-trace-ns
    "W" 'cider-eval-region
    "X" 'spacemacs/cider-display-error-buffer
    "Y" 'spacemacs/cider-test-rerun-failed-tests
    "Z" 'cljr-add-project-dependency
    "a" 'cider-apropos
    "b" 'cider-eval-buffer
    "c" 'spacemacs/cider-send-function-to-repl
    ;; "d" RESERVED
    ;; "e" RESERVED
    ;; "f" RESERVED
    ;; "g" RESERVED
    ;; "h" RESERVED
    "i" 'cider-eval-last-sexp
    "j" 'my-eval-and-next
    "k" 'my-eval-and-previous
    "l" 'spacemacs/cider-send-last-sexp-to-repl
    "m" 'cider-macroexpand-1
    "n" 'cider-repl-set-ns
    "o" 'cider-eval-last-sexp-and-replace
    "p" 'cider-test-run-project-tests
    "q" 'cider-refresh
    ;; "r" RESERVED
    ;; "s" RESERVED
    ;; "t" RESERVED
    "u" 'spacemacs/cider-test-run-focused-test
    "v" 'cider-toggle-trace-var
    "w" 'spacemacs/cider-send-region-to-repl
    "x" 'cider-debug-defun-at-point
    "y" 'cider-test-run-ns-tests
    "z" 'cljr-add-require-to-ns))

;; SPECIAL BUFFERS

(use-package cider
  :defer t
  :init
  (progn
    ;; stacktrace mode
    (define-key cider-stacktrace-mode-map "J" 'cider-test-next-result)
    (define-key cider-stacktrace-mode-map "K" 'cider-test-previous-result)
    (evilified-state-evilify cider-stacktrace-mode cider-stacktrace-mode-map
      (kbd "C-k") 'tmux-nav-up
      (kbd "C-j") 'tmux-nav-down)
    ;; test report mode
    (define-key cider-test-report-mode-map "J" 'cider-test-next-result)
    (define-key cider-test-report-mode-map "K" 'cider-test-previous-result)
    (evilified-state-evilify cider-test-report-mode cider-test-report-mode-map
      (kbd "C-k") 'tmux-nav-up
      (kbd "C-j") 'tmux-nav-down)))
