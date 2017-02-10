                                        ; CONFIG

(setq org-startup-folded 'children)

;; plant-uml
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))
(setq org-plantuml-jar-path (expand-file-name "~/bin/plantuml.jar"))
(setq org-confirm-babel-evaluate nil)

                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'org-mode "hh" 'org-html-export-to-html)
