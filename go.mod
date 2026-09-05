module github.com/softwaresalt/intercom-go

go 1.24

// toolchain pinned above the go.mod language floor to remediate GO-2025-3750
// (Inconsistent handling of O_CREATE|O_EXCL on Unix and Windows in os/syscall,
// fixed in go1.23.10+). The 1.22.x release line was never patched for this
// CVE, so govulncheck cannot pass on any 1.22.x toolchain. The language floor
// itself was raised to Go 1.24 (shipment 004-S / U-A1) to satisfy the
// github.com/github/copilot-sdk/go dependency's declared floor ahead of its
// phase-C2 adoption; the toolchain pin above remains newer than the floor.
toolchain go1.26.5

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/github/copilot-sdk/go v1.0.11
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/coder/websocket v1.8.15 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
)
