// Package main is the entry point for the dot command-line tool.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	// Local package
	"dot"
)

type appRunner interface {
	Run(context.Context, []string) error
}

func run(ctx context.Context, app appRunner, args []string, stderr io.Writer) int {
	if err := app.Run(ctx, args); err != nil {
		if errors.Is(err, context.Canceled) {
			return 130 // 128 + SIGINT, the conventional interrupt exit code
		}
		_, _ = fmt.Fprintln(stderr, "dot:", err)
		return 1
	}
	return 0
}

func main() {
	// Cancel the root context on interrupt so in-flight operations stop and deferred
	// cleanup runs (e.g. removing chezmoi's temporary probe files) instead of a hard kill.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx, dot.NewApp(), os.Args, os.Stderr))
}
