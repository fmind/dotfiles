// Package main is the entry point for the dot command-line tool.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	// Local package
	"dot"
)

func main() {
	// Cancel the root context on interrupt so in-flight operations stop and deferred
	// cleanup runs (e.g. removing chezmoi's temporary probe files) instead of a hard kill.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := dot.NewApp()
	if err := app.Run(ctx, os.Args); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(130) // 128 + SIGINT, the conventional interrupt exit code
		}
		_, _ = fmt.Fprintln(os.Stderr, "dot:", err)
		os.Exit(1)
	}
}
