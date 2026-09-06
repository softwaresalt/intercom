package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
)

// Validate performs one unconditional, fail-fast validation pass over c,
// returning on the first violated rule. Rule order is fixed so a given
// invalid config always yields the same error (requirement R8).
//
// Canonicalized paths (the default workspace root, and each non-empty
// [[workspace]].path) are computed into locals and committed to the
// receiver only after every rule has passed, so a failed validation never
// mutates its receiver (Decision R10, Protected Invariant I6).
func (c *Config) Validate() (Report, error) {
	var report Report

	// Rule 1: max_concurrent_sessions must be non-zero (config.rs:540).
	if c.MaxConcurrentSessions == 0 {
		return report, apperr.New(apperr.KindConfig, "max_concurrent_sessions must be greater than zero")
	}

	// Rule 2a: default_workspace_root must be set. Required so an absent
	// or explicitly-empty root never reaches pathsafe.NewRoot(""), where
	// filepath.Abs("") resolves to the process working directory and
	// EvalSymlinks succeeds — silently binding the workspace root to the
	// CWD. This is a Go-specific hazard with no oracle analogue (Rust's
	// canonicalize("") returns ENOENT).
	if c.DefaultWorkspaceRoot == "" {
		return report, apperr.New(apperr.KindConfig, "default_workspace_root must be set")
	}

	// Rule 2b: default_workspace_root must canonicalize (config.rs:543-547).
	root, err := pathsafe.NewRoot(c.DefaultWorkspaceRoot)
	if err != nil {
		return report, apperr.Newf(apperr.KindConfig, "default_workspace_root invalid: %s", pathErrorMessage(err))
	}

	// Rules 3-4: per-entry workspace_id presence and uniqueness checks.
	if err := validateMappingFields(c.Workspaces); err != nil {
		return report, err
	}

	// Rule 5: copilot.cli_path must be well-formed.
	if err := validateCLIPath(c.Copilot.CLIPath, &report); err != nil {
		return report, err
	}

	// Rule 6: each non-empty [[workspace]].path must canonicalize
	// (divergence V7 — the oracle passes this straight into a subprocess
	// current_dir() unvalidated; P2 treats each mapping path as an
	// explicitly authorized workspace root, closing a containment gap a
	// later phase would otherwise inherit).
	canonicalWorkspacePaths := make([]string, len(c.Workspaces))
	for i, m := range c.Workspaces {
		if m.Path == "" {
			continue
		}
		wsRoot, err := pathsafe.NewRoot(m.Path)
		if err != nil {
			return report, apperr.Newf(apperr.KindConfig, "workspace path invalid for workspace_id '%s': %s", m.WorkspaceID, pathErrorMessage(err))
		}
		canonicalWorkspacePaths[i] = wsRoot.Path()
	}

	// Rule 7: database.path must not contain a '..' segment (divergence
	// V8 — the oracle leaves it unrestricted while a later phase creates
	// parent directories there, an arbitrary-write primitive otherwise).
	// Absolute paths remain permitted as an explicit, visible operator
	// privilege.
	if pathsafe.ContainsDotDotSegment(c.Database.Path) {
		return report, apperr.New(apperr.KindConfig, "database.path must not contain '..' segments")
	}

	// Deferred commit: every rule has passed.
	c.DefaultWorkspaceRoot = root.Path()
	for i, p := range canonicalWorkspacePaths {
		if p != "" {
			c.Workspaces[i].Path = p
		}
	}

	return report, nil
}

// pathErrorMessage extracts the underlying message from a
// pathsafe.NewRoot error, avoiding the redundant nested phrasing that
// would result from embedding the full apperr.Error.Error() rendering
// (which already carries its own "path violation: " kind prefix) inside
// this package's own "{field} invalid: {message}" wrapper. Falls back to
// err.Error() if err is not an *apperr.Error (defensive; pathsafe.NewRoot
// always returns one today).
func pathErrorMessage(err error) string {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		return appErr.Message()
	}
	return err.Error()
}

// validateCLIPath implements validation rule 5. Empty is valid and returns
// early with no advisory. An absolute path must exist. A Windows drive-
// relative path and any other non-absolute path containing a separator are
// rejected. A bare name is permitted, but if exec.LookPath cannot resolve
// it a single non-fatal advisory is appended to report.Warnings.
func validateCLIPath(cliPath string, report *Report) error {
	if cliPath == "" {
		return nil
	}

	if filepath.IsAbs(cliPath) {
		if _, err := os.Stat(cliPath); err != nil {
			return apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' does not exist", cliPath)
		}
		return nil
	}

	if hasWindowsDrivePrefix(cliPath) {
		return apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not be drive-relative; use an absolute path, a bare name, or empty", cliPath)
	}

	if strings.ContainsAny(cliPath, `/\`) {
		return apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not be relative; use an absolute path, a bare name, or empty", cliPath)
	}

	if _, err := exec.LookPath(cliPath); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("copilot.cli_path '%s' was not resolvable via PATH", cliPath))
	}

	return nil
}

func hasWindowsDrivePrefix(path string) bool {
	if len(path) < 2 || path[1] != ':' {
		return false
	}
	b := path[0]
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

// validateMappingFields implements validation rules 3-4: each
// [[workspace]] entry must declare a non-empty workspace_id (rule 3), and
// workspace_id must be unique across entries (rule 4).
//
// Entries are checked in declaration order; the first offending entry
// wins, and within one entry rule 3 is checked before rule 4.
func validateMappingFields(mappings []WorkspaceMapping) error {
	seenWorkspaceIDs := make(map[string]struct{}, len(mappings))
	for _, m := range mappings {
		if m.WorkspaceID == "" {
			return apperr.New(apperr.KindConfig, "workspace_id cannot be empty in [[workspace]] entry")
		}
		if _, dup := seenWorkspaceIDs[m.WorkspaceID]; dup {
			return apperr.Newf(apperr.KindConfig, "duplicate workspace_id '%s' in [[workspace]] entries", m.WorkspaceID)
		}
		seenWorkspaceIDs[m.WorkspaceID] = struct{}{}
	}
	return nil
}
