(defconst my-fasd-packages '(fasd))

(defun my-fasd/post-init-fasd ()
  (use-package fasd
    :defer t
    :init
    (progn
      (spacemacs/set-leader-keys "fa" 'fasd-find-file)
      (spacemacs/set-leader-keys "fi" 'fasd-find-file-only)
      (spacemacs/set-leader-keys "fd" 'fasd-find-directory-only))))
