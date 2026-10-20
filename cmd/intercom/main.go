// Command intercom is the intercom-go server entrypoint.
package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/softwaresalt/intercom-go/internal/app"
	"github.com/softwaresalt/intercom-go/internal/config"
)

// stderr is the destination for structured logs. Overridable in tests.
var stderr io.Writer = os.Stderr

func main() {
	if err := newRootCmd().Execute(); err != nil {
		newLogger("info").Error("intercom exited with error", "error", err)
		os.Exit(1)
	}
}

// runFunc starts the process for the given options and returns when it exits.
type runFunc func(ctx context.Context, o app.Options) error

// runApp is the production runFunc: it assembles the process and runs it until
// the local TUI exits or an interrupt/terminate signal arrives, then tears it
// down.
func runApp(ctx context.Context, o app.Options) error {
	a, err := app.Start(ctx, o)
	if err != nil {
		return err
	}
	select {
	case <-a.Done():
	case <-o.Signals:
	}
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownBudget)
	defer cancel()
	return a.Stop(sctx)
}

// shutdownBudget bounds the teardown at exit.
const shutdownBudget = 30 * time.Second

// newRootCmd builds the intercom root command with the production runner.
func newRootCmd() *cobra.Command { return newRootCmdWith(runApp) }

// newRootCmdWith builds the intercom root command: the flag surface
// (workspace, config path, port, log level) and the structured-logging
// bootstrap. run assembles and runs the process.
func newRootCmdWith(run runFunc) *cobra.Command {
	var configPath, workspace, logLevel string
	var port uint16

	cmd := &cobra.Command{
		Use:           "intercom",
		Short:         "intercom server entrypoint",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger := newLogger(logLevel)
			logger.Info("intercom starting", "config", configPath)

			sig := make(chan os.Signal, 2)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(sig)

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return run(ctx, app.Options{
				ConfigPath: configPath,
				Workspace:  workspace,
				Port:       port,
				PortSet:    cmd.Flags().Changed("port"),
				NewAgent:   app.CopilotAgent,
				Logger:     logger,
				Signals:    sig,
			})
		},
	}

	cmd.Flags().StringVar(&configPath, "config", config.DefaultConfigPath, "path to the intercom config file")
	cmd.Flags().StringVar(&workspace, "workspace", "", "workspace directory (default: the config's default_workspace_root)")
	cmd.Flags().Uint16Var(&port, "port", 0, "loopback TCP port for the WebSocket listener (default: the config's http_port; 0 picks a free port)")
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
