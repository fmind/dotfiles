package dot

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/urfave/cli/v3"
)

// Version is the base semantic version. It can be overridden at build time via
// -ldflags "-X dot.Version=...", and is enriched with VCS metadata by VersionString.
var Version = "1.4.0"

// NewVersionCmd constructs the top-level version command.
func NewVersionCmd(state *GlobalState) *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: []string{"n"},
		Usage:   "Print the version and embedded build metadata",
		Action: func(_ context.Context, _ *cli.Command) error {
			_, _ = fmt.Fprintln(state.Stdout, VersionString())
			return nil
		},
	}
}

// VersionString returns the base version enriched with the VCS revision and dirty
// flag that the Go toolchain embeds automatically for builds from a Git checkout,
// so a user can tell whether an installed binary matches their current sources.
func VersionString() string {
	base := "dot " + Version
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return base
	}

	var revision string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}

	if revision == "" {
		return base
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if dirty {
		revision += ", dirty"
	}
	return base + " (" + revision + ")"
}
