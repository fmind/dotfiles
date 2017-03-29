                                        ; CONFIGURATION

;; plant-uml
(setq org-plantuml-jar-path (expand-file-name "~/bin/plantuml.jar"))
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))
