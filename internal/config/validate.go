package config

import (
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

	// Rule 9: each non-empty [[workspace]].path must canonicalize
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
			return report, apperr.Newf(apperr.KindConfig, "workspace path invalid for workspace_id '%s': %s", m.WorkspaceID, err.Error())
		}
		canonicalWorkspacePaths[i] = wsRoot.Path()
	}

	// Rule 10: database.path must not contain a '..' segment (divergence
	// V8 — the oracle leaves it unrestricted while a later phase creates
	// parent directories there, an arbitrary-write primitive otherwise).
	// Absolute paths remain permitted as an explicit, visible operator
	// privilege.
	if containsDotDotSegment(c.Database.Path) {
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

// containsDotDotSegment reports whether path contains a literal ".."
// path-separator-delimited segment, checked against the original
// (uncleaned) string so a traversal component cannot be hidden by
// filepath.Clean's collapsing behavior. Both '/' and '\\' are treated as
// separators regardless of GOOS, since a config file is portable text
// that may be authored on a different platform than the one that loads
// it.
func containsDotDotSegment(path string) bool {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		if part == ".." {
			return true
		}
	}
	return false
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
