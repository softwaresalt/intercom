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
func (c *Config) WorkspaceRootForChannel(channelID string) string {
	if m, ok := c.ResolveWorkspaceByChannelID(channelID); ok && m.Path != "" {
		return m.Path
	}
	return c.DefaultWorkspaceRoot
}
