package config

import (
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
		return report, apperr.Newf(apperr.KindConfig, "default_workspace_root invalid: %s", err.Error())
	}

	// Deferred commit: every rule implemented so far has passed.
	c.DefaultWorkspaceRoot = root.Path()

	return report, nil
}
