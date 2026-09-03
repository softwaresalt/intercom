// Package pathsafe implements the workspace path-containment security
// control ported from the agent-intercom behavioral oracle
// (softwaresalt/agent-intercom @ 41df772, src/diff/path_safety.rs). Brief
// section 10 marks workspace isolation NON-NEGOTIABLE.
//
// Known limitations (both oracle-parity, not silently inherited):
//   - TOCTOU: there is a window between path validation (Resolve) and actual
//     filesystem use during which the filesystem may change.
//   - EvalSymlinks does not resolve hardlinks, so a pre-existing in-workspace
//     hardlink to an external file on the same volume passes validation
//     (finding SEC-5).
package pathsafe

import (
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// uncPrefix is the volume-path prefix Go's filepath.EvalSymlinks emits on
// Windows via GetFinalPathNameByHandle.
const uncPrefix = `\\?\`

// Root is a canonicalized workspace root. Construct with NewRoot; the zero
// value is not valid.
type Root struct {
	path string
}

// Path returns the canonicalized, absolute root path.
func (r Root) Path() string {
	return r.path
}

// NewRoot canonicalizes dir exactly once: filepath.Abs, then
// filepath.EvalSymlinks, then stripUNCPrefix. Canonicalizing once amortizes
// what would otherwise be an EvalSymlinks syscall walk on every Resolve
// call — P10's diff-apply may validate 10-100 paths against one root
// (finding GO-4).
func NewRoot(dir string) (Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Root{}, apperr.New(apperr.KindPathViolation, "workspace root invalid: "+err.Error())
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return Root{}, apperr.New(apperr.KindPathViolation, "workspace root invalid: "+err.Error())
	}

	return Root{path: stripUNCPrefix(resolved)}, nil
}

// stripUNCPrefix removes a leading \\?\ prefix, which Go's EvalSymlinks
// emits on Windows.
//
// Provenance note: this mirrors src/config.rs:24 strip_unc_prefix, which the
// oracle applies to default_workspace_root at src/config.rs:546 — it is NOT
// called from src/diff/path_safety.rs, which relies on Rust's
// component-aware Path::starts_with instead. The stripping is therefore a
// Go-specific adaptation required because Go's containment comparison is
// string-based, not an oracle-faithful port of path_safety.rs (findings
// SEC-4 / ARCH-4).
func stripUNCPrefix(p string) string {
	return strings.TrimPrefix(p, uncPrefix)
}
