                                        ; KEYBINDINGS

(spacemacs/set-leader-keys-for-major-mode 'python-mode
  "b" 'python-shell-send-buffer
  "B" 'python-shell-send-buffer-switch
  "f" 'python-shell-send-defun
  "F" 'python-shell-send-defun-switch
  "i" 'python-shell-send-region
  "I" 'python-shell-send-region-switch)
