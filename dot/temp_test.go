package dot

import (
	"strings"
	"testing"
)

func TestRemoveTemporaryDirectoryReportsErrors(t *testing.T) {
	err := removeTemporaryDirectory("invalid\x00path", "test workspace")
	if err == nil || !strings.Contains(err.Error(), "failed to remove temporary test workspace") {
		t.Fatalf("expected temporary-directory cleanup error, got %v", err)
	}
}
