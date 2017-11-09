                                        ; CONFIGURATION

;; agenda
(setq org-agenda-files (list "~/org/home.org"
                             "~/org/work.org" 
                             "~/org/idea.org"
                             "~/org/tool.org"))

;; plant-uml
(setq org-plantuml-jar-path (expand-file-name "~/bin/plantuml.jar"))
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; HOOKS

(add-hook 'org-mode-hook 'turn-on-auto-fill)
