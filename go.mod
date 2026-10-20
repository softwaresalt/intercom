module github.com/softwaresalt/intercom-go

go 1.25.0

// toolchain pinned above the go.mod language floor to remediate GO-2025-3750
// (Inconsistent handling of O_CREATE|O_EXCL on Unix and Windows in os/syscall,
// fixed in go1.23.10+). The 1.22.x release line was never patched for this
// CVE, so govulncheck cannot pass on any 1.22.x toolchain. The language floor
// itself was raised to Go 1.24 (shipment 004-S / U-A1) to satisfy the
// github.com/github/copilot-sdk/go dependency's declared floor ahead of its
// phase-C2 adoption; the toolchain pin above remains newer than the floor.
// The toolchain was later raised from go1.26.5 to go1.26.9 to remediate stdlib
// advisories GO-2026-6603, GO-2026-6607, GO-2026-6608, GO-2026-6611,
// GO-2026-6612, GO-2026-6613, and GO-2026-6617 (net/http, crypto/tls,
// net/textproto, mime/multipart; fixed in go1.26.9). This is a toolchain-only
// change: the go 1.24 language floor was unchanged by that bump (see
// docs/decisions/2026-09-04-go-toolchain-pin-maintenance-note.md).
// The language floor was later raised from go 1.24 to go 1.25.0 (shipment
// 061-S / 069-F) because the TUI dependencies' security-fixed transitive
// modules (golang.org/x/sys v0.44.0 for GO-2026-5024, golang.org/x/text v0.41.0
// for GO-2026-6629 and GO-2026-5970) declare go 1.25.0; the toolchain pin above
// is unchanged and still newer.
toolchain go1.26.9

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/charmbracelet/bubbles v1.0.0
	github.com/charmbracelet/bubbletea v1.3.10
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/coder/websocket v1.8.15
	github.com/github/copilot-sdk/go v1.0.11
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.4.1 // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.9.0 // indirect
	github.com/clipperhouse/stringish v0.1.1 // indirect
	github.com/clipperhouse/uax29/v2 v2.5.0 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.41.0 // indirect
)
