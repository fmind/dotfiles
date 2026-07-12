package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// asset is a vendored third-party file pinned to an exact version and content
// hash. The runtime never touches a CDN — assets are embedded via go:embed and
// served from /static/ — so this build-time step is the only network access,
// and the sha256 check fails loudly if an upstream artifact ever changes under
// a pinned URL (typo, upstream re-tag, or CDN compromise).
type asset struct {
	name   string
	url    string
	sha256 string
}

// Bump a version by editing its url and sha256 together. On a mismatch the error
// prints the actual sum, so paste that in after a deliberate upgrade.
var assets = []asset{
	{
		name:   "htmx.min.js",
		url:    "https://unpkg.com/htmx.org@2.0.10/dist/htmx.min.js",
		sha256: "71ea67185bfa8c98c39d31717c6fce5d852370fcdfd129db4543774d3145c0de",
	},
	{
		name:   "alpine.min.js",
		url:    "https://unpkg.com/alpinejs@3.15.12/dist/cdn.min.js",
		sha256: "57b37d7cae9a27d965fdae4adcc844245dfdc407e655aee85dcfff3a08036a3f",
	},
}

func main() {
	staticDir := filepath.Join("static", "vendor")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		fmt.Printf("Error creating static directory: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for _, a := range assets {
		dest := filepath.Join(staticDir, a.name)
		fmt.Printf("Vendoring %s from %s...\n", a.name, a.url)
		if err := vendor(ctx, dest, a); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Vendored assets verified and written.")
}

// vendor downloads a.url, verifies its sha256, and writes it to dest only when
// the hash matches — a mismatch aborts without touching the file on disk.
func vendor(ctx context.Context, dest string, a asset) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", a.url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for %s: %s", a.url, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading %s: %w", a.url, err)
	}

	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != a.sha256 {
		return fmt.Errorf("sha256 mismatch for %s:\n  want %s\n  got  %s", a.name, a.sha256, got)
	}

	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", dest, err)
	}
	return nil
}
