(setq PRIVATE "~/.spacemacs.d")
;; BACKUP {{{
(setq create-lockfiles nil)
(add-hook 'focus-out-hook (lambda () (save-some-buffers t)))
;;}}}
;; PROJECTILE {{{
(setq projectile-globally-ignored-directories '("out" "target"))
(setq projectile-globally-ignored-file-suffixes '("jpg" "png" "gif" "pyc"))
;;}}}
;; ABBREVIATIONS {{{
(setq-default abbrev-mode t)
(setq save-abbrevs 'silently)
(setq abbrev-file-name (concat PRIVATE "abbreviations"))
;;}}}
;; INITIALIZATION {{{
(setq vc-follow-symlinks t)
(add-hook 'after-init-hook 'global-company-mode)
;;}}}
