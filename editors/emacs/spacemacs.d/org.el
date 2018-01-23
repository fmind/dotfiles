                                        ; CONF

;; plant-uml
(setq org-plantuml-jar-path "/usr/share/java/plantuml.jar")
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; KEYS

(spacemacs/set-leader-keys-for-major-mode 'org-mode
  "r" 'org-reveal-export-to-html)
