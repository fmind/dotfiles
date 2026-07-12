package dot

import (
	"errors"
	"fmt"
	"os"
)

func removeTemporaryFile(path, purpose string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("failed to remove temporary %s %s: %w", purpose, path, err)
}

func removeTemporaryDirectory(path, purpose string) error {
	err := os.RemoveAll(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("failed to remove temporary %s %s: %w", purpose, path, err)
}
