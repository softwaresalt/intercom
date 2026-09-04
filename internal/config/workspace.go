package config

// WorkspaceRoot returns the workspace root directory for workspaceID: the
// matching mapping's own Path if one is set, otherwise DefaultWorkspaceRoot.
// Safe to call on a Config with no [[workspace]] entries at all. If more than
// one entry declares the same workspace_id (prevented by validation rule 4 on
// a Validate()d config, but not assumed here), the first match in declaration
// order wins.
//
// Precondition: the returned path is only guaranteed to be canonicalized and
// contained within an authorized workspace root when c has passed Validate()
// (i.e. was produced by Load or Decode with a nil error). Calling this on a
// *Config built directly by struct literal, or on one for which Validate()
// returned a non-nil error, may return a raw, non-canonicalized Path that
// looks identical to a validated one. A caller must not treat the returned
// string as containment-safe otherwise.
func (c *Config) WorkspaceRoot(workspaceID string) string {
	for _, m := range c.Workspaces {
		if m.WorkspaceID == workspaceID {
			if m.Path != "" {
				return m.Path
			}
			break
		}
	}
	return c.DefaultWorkspaceRoot
}
