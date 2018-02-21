(defun my-defonce-toggle ()
  (interactive)
  (let* ((line (thing-at-point 'line t))
         (line (cond
                ((string-match-p "^(def " line) (replace-regexp-in-string "(def " "(defonce " line))
                ((string-match-p "^(defonce " line) (replace-regexp-in-string "(defonce " "(def " line))
                (t line))))
    (kill-whole-line) (insert line) (previous-line)))
