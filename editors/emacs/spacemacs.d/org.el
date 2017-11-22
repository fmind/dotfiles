                                        ; CONF

;; plant-uml
(setq org-plantuml-jar-path (expand-file-name "~/bin/plantuml.jar"))
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; FUNS

(defun my-org-html-export-body ()
  (interactive)
  (org-html-export-to-html nil nil nil t nil)
  (message (concat "Exporting buffer to " (buffer-name) " (html)")))

                                        ; HOOK

(add-hook 'org-mode-hook 'turn-on-auto-fill)
(add-hook 'org-mode-hook 'spacemacs/toggle-highlight-current-line-globally-on)

                                        ; KEYS

(spacemacs/set-leader-keys-for-major-mode 'org-mode
  "w" 'my-org-html-export-body)
