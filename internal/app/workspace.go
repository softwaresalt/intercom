package app

import (
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
)

// resolveWorkspace returns the canonical workspace root for the process.
//
// An empty flag selects the config's default_workspace_root. A non-empty flag
// must name an existing directory that lies inside one of the authorized roots
// (default_workspace_root or a [[workspace]].path, all canonicalized by
// config validation). The check is made through pathsafe: the flag is
// canonicalized with [pathsafe.NewRoot] (resolving symlinks and Windows
// reparse points, so a link out of a root canonicalizes to its real, outside
// location), and the part below the authorized root is resolved with
// [pathsafe.Root.Resolve], the containment control for workspace-relative
// paths. Anything else is refused (fail closed).
func resolveWorkspace(cfg *config.Config, flag string) (string, error) {
	roots := make([]string, 0, 1+len(cfg.Workspaces))
	roots = append(roots, cfg.DefaultWorkspaceRoot)
	for _, m := range cfg.Workspaces {
		if m.Path != "" {
			roots = append(roots, m.Path)
		}
	}

	if flag == "" {
		base, err := pathsafe.NewRoot(cfg.DefaultWorkspaceRoot)
		if err != nil {
			return "", err
		}
		return base.Path(), nil
	}

	target, err := pathsafe.NewRoot(flag)
	if err != nil {
		return "", err
	}
	for _, r := range roots {
		base, err := pathsafe.NewRoot(r)
		if err != nil {
			return "", err
		}
		rel, err := filepath.Rel(base.Path(), target.Path())
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue // a different volume, or not below this root
		}
		if rel == "." {
			return base.Path(), nil
		}
		return base.Resolve(rel)
	}
	return "", apperr.New(apperr.KindPathViolation, "workspace is outside every authorized workspace root")
}
