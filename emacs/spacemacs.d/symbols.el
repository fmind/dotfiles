                                        ; FUNCTIONS

(defun my-insert-lambda ()
  (interactive)
  (insert-char (make-char 'greek-iso8859-7 107) 1))
(put 'my 'delete-selection t)

                                        ; KEYBINDINGS

(spacemacs/set-leader-keys "ol" 'my-insert-lambda)
