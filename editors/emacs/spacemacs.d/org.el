                                        ; CONF

;; plant-uml
(setq plantuml-jar-path "/usr/share/java/plantuml.jar")
(setq org-plantuml-jar-path "/usr/share/java/plantuml.jar")
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; HOOK

(add-hook 'org-mode-hook 'spacemacs/toggle-visual-line-navigation-on)

                                        ; KEYS

(spacemacs/set-leader-keys-for-major-mode 'org-mode
  "r" 'org-reveal-export-to-html)
