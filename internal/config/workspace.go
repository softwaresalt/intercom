package config

// ResolveChannelID returns the channel_id mapped to workspaceID, and
// whether a mapping was found. If more than one entry declares the same
// workspace_id (prevented by validation rule 5 on a Validate()d config,
// but not assumed here), the first match in declaration order wins.
func (c *Config) ResolveChannelID(workspaceID string) (string, bool) {
	for _, m := range c.Workspaces {
		if m.WorkspaceID == workspaceID {
			return m.ChannelID, true
		}
	}
	return "", false
}

// ResolveWorkspaceByChannelID returns the WorkspaceMapping whose
// channel_id equals channelID, and whether one was found. It returns a
// value, not a pointer, so a caller cannot mutate a resolved mapping in
// place or retain a stale alias into a slice a future hot-reload might
// replace. The first match in declaration order wins.
//
// Precondition: the containment guarantee on the returned mapping's Path
// (that it is absolute, symlink-resolved, and canonicalized within an
// authorized workspace root — divergence V7) holds only if c was produced
// by Load or Decode and its returned error was nil (i.e. c has passed
// Validate()). Calling this on a *Config built directly by struct literal,
// or on one for which Validate() returned a non-nil error, may return a
// raw, non-canonicalized Path that looks identical to a validated one.
func (c *Config) ResolveWorkspaceByChannelID(channelID string) (WorkspaceMapping, bool) {
	for _, m := range c.Workspaces {
		if m.ChannelID == channelID {
			return m, true
		}
	}
	return WorkspaceMapping{}, false
}

// WorkspaceRootForChannel returns the workspace root directory for
// channelID: the mapping's own Path if one is set, otherwise
// DefaultWorkspaceRoot. Safe to call on a Config with no [[workspace]]
// entries at all.
//
// Precondition: as with ResolveWorkspaceByChannelID, the returned path is
// only guaranteed to be canonicalized and contained within an authorized
// workspace root when c has passed Validate() (i.e. was produced by Load
// or Decode with a nil error). A caller must not treat the returned
// string as containment-safe otherwise.
func (c *Config) WorkspaceRootForChannel(channelID string) string {
	if m, ok := c.ResolveWorkspaceByChannelID(channelID); ok && m.Path != "" {
		return m.Path
	}
	return c.DefaultWorkspaceRoot
}
