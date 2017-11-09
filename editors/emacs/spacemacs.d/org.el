                                        ; CONF

;; plant-uml
(setq org-plantuml-jar-path (expand-file-name "~/bin/plantuml.jar"))
(org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))

                                        ; HOOK

(add-hook 'org-mode-hook 'turn-on-auto-fill)
(add-hook 'org-mode-hook 'spacemacs/toggle-highlight-current-line-globally-on)
