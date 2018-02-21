(defconst my-config-packages '(golden-ratio))

(defun my-config/post-init-golden-ratio ()
  (use-package golden-ratio
      :defer t
      :init
      (progn
        (golden-ratio-mode))))
