// Package copilotprobe is a DISPOSABLE phase-C2 proving spike for the pinned
// github.com/github/copilot-sdk/go dependency (shipment 005-S, sub-epic B).
//
// This package produces NO production SDK wiring and NO anti-corruption
// layer. It exists solely to prove D2's three required surfaces (S1
// permission round-trip, S2 event-union handling, S3 cancellation/shutdown)
// and to answer the four design-mandated C2 spike questions (SQ-a..SQ-d,
// see docs/design-docs/intercom-go-backend-architecture.md §§3.1, 4.2, 5.1,
// 5.4). See docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md and
// docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md.
//
// Both binaries (cmd/intercom, cmd/intercom-ctl) remain unmodified and keep
// returning their "not implemented" sentinel; nothing in this package is
// imported by production code.
package copilotprobe

import (
	"context"

	copilot "github.com/github/copilot-sdk/go"
)

// NewClient constructs a Copilot SDK client via the SDK's public API. This
// is the package's genuine, non-test importer required by D9a so that
// `go mod tidy` does not prune the pinned dependency (an ordinary,
// non-build-tagged file — see the plan's "Untagged importer" decision).
//
// It applies no probe-specific behaviour: callers configure ClientOptions
// (working directory, permission handler, etc.) for their own scenario.
// This function performs no I/O; the caller must still call Start.
func NewClient(opts *copilot.ClientOptions) *copilot.Client {
	return copilot.NewClient(opts)
}

// StartClient is a small convenience wrapper proving the SDK's documented
// Start/Stop lifecycle via its public API. Callers are responsible for
// calling the returned stop function (typically via defer) exactly once.
func StartClient(ctx context.Context, opts *copilot.ClientOptions) (*copilot.Client, func() error, error) {
	client := NewClient(opts)
	if err := client.Start(ctx); err != nil {
		return nil, nil, err
	}
	return client, client.Stop, nil
}
