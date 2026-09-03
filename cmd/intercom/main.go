// Command intercom is the intercom-go server entrypoint.
package main

import (
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// errNotImplemented is returned by RunE until server behavior is implemented.
var errNotImplemented = errors.New("intercom: not implemented")

// stderr is the destination for structured logs. Overridable in tests.
var stderr io.Writer = os.Stderr

func main() {
	if err := newRootCmd().Execute(); err != nil {
		newLogger("info").Error("intercom exited with error", "error", err)
		os.Exit(1)
	}
}

// newRootCmd builds the intercom root command with its flag surface and
// structured-logging bootstrap. RunE returns a not-implemented sentinel
// until server behavior lands in a later slice.
func newRootCmd() *cobra.Command {
	var configPath string
	var logLevel string

	cmd := &cobra.Command{
		Use:           "intercom",
		Short:         "intercom server entrypoint",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			logger := newLogger(logLevel)
			logger.Info("intercom starting", "config", configPath)
			return errNotImplemented
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to the intercom config file")
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
