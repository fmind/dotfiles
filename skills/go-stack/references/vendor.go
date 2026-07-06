package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	staticDir := filepath.Join("static", "vendor")
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		fmt.Printf("Error creating static directory: %v\n", err)
		os.Exit(1)
	}

	assets := map[string]string{
		"htmx.min.js":   "https://unpkg.com/htmx.org@2.0.10/dist/htmx.min.js",
		"alpine.min.js": "https://unpkg.com/alpinejs@3.15.12/dist/cdn.min.js",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for filename, url := range assets {
		dest := filepath.Join(staticDir, filename)
		fmt.Printf("Downloading %s to %s...\n", url, dest)
		if err := downloadFile(ctx, dest, url); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Vendored assets downloaded successfully.")
}

func downloadFile(ctx context.Context, dest, url string) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for %s: %s", url, resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}
