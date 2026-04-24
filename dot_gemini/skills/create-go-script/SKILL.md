---
name: create-go-script
description: Generate clean, standalone Go programs — single-file mains with cobra and slog.
---

# Create Go Script

This skill guides you in producing small, idiomatic Go programs that read like a script but compile to a single static binary.

## Project Layout

For anything beyond ~100 LOC, use a real Go module:

```text
mytool/
├── go.mod
├── go.sum
└── main.go
```

Bootstrap with:

```bash
mkdir mytool && cd mytool
go mod init github.com/<user>/mytool
go get github.com/spf13/cobra@latest
```

For one-off snippets, a single `main.go` with `go run main.go` is fine.

## Template (`main.go`)

```go
package main

import (
    "context"
    "errors"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    var verbose bool
    cmd := &cobra.Command{
        Use:           "mytool",
        Short:         "One-line description.",
        SilenceUsage:  true,
        SilenceErrors: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            level := slog.LevelInfo
            if verbose {
                level = slog.LevelDebug
            }
            slog.SetDefault(slog.New(slog.NewTextHandler(
                os.Stderr, &slog.HandlerOptions{Level: level})))

            return run(cmd.Context(), args)
        },
    }
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "debug logging")

    if err := cmd.ExecuteContext(ctx); err != nil {
        slog.Error("fatal", "err", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, args []string) error {
    if len(args) == 0 {
        return errors.New("at least one argument required")
    }
    slog.Debug("starting", "args", args)
    // ... implementation ...
    return nil
}
```

## Core Principles

1. **`context.Context` everywhere.** Honour cancellation; wire it through
   every I/O call.
1. **`log/slog` for logs.** Stderr, structured. No `fmt.Println` for logs.
1. **Errors are values.** Return them; wrap with `%w`; let `main` log.
1. **Cobra for CLIs.** Even tiny programs benefit from `--help` and flags.
1. **Static binary.** `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w"`.
1. **Tests live next door.** `main_test.go` with table-driven `TestRun`.

## AI Agent Instructions

- Always run `go mod tidy` after editing imports.
- Run `gofmt -w .` and `golangci-lint run` before declaring victory.
- For HTTP services, prefer the standard library `net/http` over frameworks
  unless the project already pulls one in.
- For concurrency, prefer `errgroup.WithContext` over hand-rolled `sync`
  boilerplate.
