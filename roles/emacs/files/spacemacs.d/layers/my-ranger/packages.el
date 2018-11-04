(defconst my-ranger-packages '(ranger))

(defun my-ranger/post-init-ranger ()
  (use-package ranger
    :defer t
    :init
    (progn
      (ranger-override-dired-mode t))))
