// Command intercom-ctl is the companion CLI for the intercom server.
package main

import (
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// errNotImplemented is returned by RunE until control-plane behavior is
// implemented.
var errNotImplemented = errors.New("intercom-ctl: not implemented")

// stderr is the destination for structured logs. Overridable in tests.
var stderr io.Writer = os.Stderr

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd builds the intercom-ctl root command with its flag surface and
// structured-logging bootstrap. RunE returns a not-implemented sentinel
// until control-plane behavior lands in a later slice.
func newRootCmd() *cobra.Command {
	var socketPath string
	var logLevel string

	cmd := &cobra.Command{
		Use:           "intercom-ctl",
		Short:         "intercom-ctl companion CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := newLogger(logLevel)
			logger.Info("intercom-ctl starting", "socket", socketPath)
			return errNotImplemented
		},
	}

	cmd.Flags().StringVar(&socketPath, "socket", "", "path to the intercom control socket")
	cmd.Flags().StringVar(&logLevel, "log-level", "info", "log level: debug, info, warn, error")

	return cmd
}

// newLogger returns a JSON slog.Logger writing to stderr at the given level.
func newLogger(level string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: parseLevel(level)}))
}

// parseLevel maps a log-level flag value to a slog.Level, defaulting to Info
// for unrecognized values.
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
