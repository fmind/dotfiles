//go:build !linux && !darwin

package dot

import "os/exec"

func isolateProcessGroup(_ *exec.Cmd) {}
