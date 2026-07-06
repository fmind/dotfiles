package dot

import "errors"

var (
	ErrGcloudNotInstalled  = errors.New("gcloud CLI is not installed or not in PATH")
	ErrGhNotInstalled      = errors.New("gh CLI is not installed or not in PATH")
	ErrChezmoiNotInstalled = errors.New("chezmoi CLI is not installed or not in PATH")
	ErrClaspNotInstalled   = errors.New("clasp CLI is not installed or not in PATH")
	ErrGitNotInstalled     = errors.New("git CLI is not installed or not in PATH")
	ErrGwsNotInstalled     = errors.New("gws CLI is not installed or not in PATH")
	ErrNotGitRepository    = errors.New("not a git repository")
	ErrToolNotInstalled    = errors.New("required tool not installed or not in PATH")
)
