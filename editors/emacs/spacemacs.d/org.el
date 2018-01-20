                                        ; CONF

;; plant-uml
(setq plantuml-jar-path "/usr/share/java/plantuml.jar")
(setq org-plantuml-jar-path "/usr/share/java/plantuml.jar")
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; FUNS

(defun my-org-html-export-body ()
  (interactive)
  (org-html-export-to-html nil nil nil t nil)
  (message (concat "Exporting buffer to " (buffer-name) " (html)")))

                                        ; KEYS

(spacemacs/set-leader-keys-for-major-mode 'org-mode
  "w" 'my-org-html-export-body)
