module github.com/softwaresalt/intercom-go

go 1.22

// toolchain pinned above the go.mod language floor to remediate GO-2025-3750
// (Inconsistent handling of O_CREATE|O_EXCL on Unix and Windows in os/syscall,
// fixed in go1.23.10+). The 1.22.x release line was never patched for this
// CVE, so govulncheck cannot pass on any 1.22.x toolchain; language
// compatibility stays at Go 1.22 (workspace floor), only the toolchain used
// to build/test is pinned newer.
toolchain go1.26.5

require github.com/spf13/cobra v1.10.2

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)
