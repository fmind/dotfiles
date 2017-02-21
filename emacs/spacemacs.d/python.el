                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'python-mode
  "G" 'spacemacs/jump-to-definition
  "C" 'spacemacs/python-execute-file
  "D" 'spacemacs/python-toggle-breakpoint
  "I" 'spacemacs/python-remove-unused-imports
  "," 'python-shell-send-buffer
  "b" 'python-shell-send-buffer
  "B" 'python-shell-send-buffer-switch
  "f" 'python-shell-send-defun
  "F" 'python-shell-send-defun-switch
  "n" 'python-shell-send-region
  "N" 'python-shell-send-region-switch)

                                        ; NOTEBOOKS

(setq ein:use-auto-complete t)
(spacemacs/set-leader-keys "oN" 'ein:notebooklist-open)
