;; package configurations

(defconst org-brain-packages '(org-brain))

(defun org-brain/init-org-brain ()
  (use-package org-brain
    :ensure t
    :defer t
    :init
    (with-eval-after-load 'evil
      (evil-set-initial-state 'org-brain-visualize-mode 'emacs))))
