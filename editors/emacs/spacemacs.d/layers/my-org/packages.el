(defconst my-org-packages '(org))

(defun my-org/post-init-org ()
  (use-package org
    :defer t
    :init
    (progn
      (org-babel-do-load-languages 'org-babel-load-languages '((plantuml . t)))
      (add-hook 'org-mode-hook 'spacemacs/toggle-visual-line-navigation-on)
      (spacemacs/set-leader-keys-for-major-mode 'org-mode
        "r" 'org-reveal-export-to-html))))
