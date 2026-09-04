package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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

	// Rules 3-5: per-entry workspace_id/channel_id/duplicate checks.
	if err := validateMappingFields(c.Workspaces); err != nil {
		return report, err
	}

	// Rules 6-7: host_cli must be set, and if absolute, must exist.
	if err := validateHostCLI(c.HostCLI, &report); err != nil {
		return report, err
	}

	// Rule 8: duplicate channel_id across [[workspace]] entries.
	if err := validateChannelUniqueness(c.Workspaces); err != nil {
		return report, err
	}

	// Deferred commit: every rule implemented so far has passed.
	c.DefaultWorkspaceRoot = root.Path()

	return report, nil
}

// validateHostCLI implements Validation Contract rules 6-7: host_cli must
// be set (rule 6, divergence V5 — the oracle's message additionally
// mentions "to use ACP mode", a retired concept), and if it is an
// absolute path, that path must exist (rule 7). A bare or relative name is
// resolved at spawn time against PATH, not here; if exec.LookPath cannot
// resolve it, a single non-fatal advisory is appended to report.Warnings.
// The oracle's additional "not in a standard install location" warning is
// deliberately not ported — the set of standard locations is undefined
// and unverifiable, and the oracle discards the result anyway
// (main.rs:139 via .ok()).
func validateHostCLI(hostCLI string, report *Report) error {
	if hostCLI == "" {
		return apperr.New(apperr.KindConfig, "host_cli must be set")
	}

	if filepath.IsAbs(hostCLI) {
		if _, err := os.Stat(hostCLI); err != nil {
			return apperr.Newf(apperr.KindConfig, "host_cli '%s' does not exist", hostCLI)
		}
		return nil
	}

	if _, err := exec.LookPath(hostCLI); err != nil {
		report.Warnings = append(report.Warnings, fmt.Sprintf("host_cli '%s' was not resolvable via PATH", hostCLI))
	}

	return nil
}

// validateMappingFields implements Validation Contract rules 3-5: each
// [[workspace]] entry must declare a non-empty workspace_id (rule 3) and
// channel_id (rule 4), and workspace_id must be unique across entries
// (rule 5). It is a separate, unexported stage from
// validateChannelUniqueness (rule 8) because the two sit on opposite
// sides of the host_cli rules (6-7) in the fixed contract order — a single
// combined helper cannot be called at one point in the sequence without
// violating that order (attempt-1 P1-b finding).
//
// Entries are checked in declaration order; the first offending entry
// wins, and within one entry rule 3 is checked before rule 4 before rule
// 5.
func validateMappingFields(mappings []WorkspaceMapping) error {
	seenWorkspaceIDs := make(map[string]struct{}, len(mappings))
	for _, m := range mappings {
		if m.WorkspaceID == "" {
			return apperr.New(apperr.KindConfig, "workspace_id cannot be empty in [[workspace]] entry")
		}
		if m.ChannelID == "" {
			return apperr.New(apperr.KindConfig, "channel_id cannot be empty in [[workspace]] entry")
		}
		if _, dup := seenWorkspaceIDs[m.WorkspaceID]; dup {
			return apperr.Newf(apperr.KindConfig, "duplicate workspace_id '%s' in [[workspace]] entries", m.WorkspaceID)
		}
		seenWorkspaceIDs[m.WorkspaceID] = struct{}{}
	}
	return nil
}

// validateChannelUniqueness implements Validation Contract rule 8:
// channel_id must be unique across all [[workspace]] entries, so ACP
// routing maps each channel_id to exactly one workspace. Divergence V4:
// the oracle enforces this at ACP session start
// (src/slack/commands.rs:731); P2 enforces it eagerly at load, since ACP
// session start does not exist until a later phase.
func validateChannelUniqueness(mappings []WorkspaceMapping) error {
	seen := make(map[string]struct{}, len(mappings))
	for _, m := range mappings {
		if _, dup := seen[m.ChannelID]; dup {
			return apperr.Newf(apperr.KindConfig, "duplicate channel_id '%s' in [[workspace]] entries; ACP routing requires each channel_id to map to exactly one workspace", m.ChannelID)
		}
		seen[m.ChannelID] = struct{}{}
	}
	return nil
}
