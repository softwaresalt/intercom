package retiredarch

import "testing"

// TestMatchesForbiddenParts_HitAndNearMiss gives each forbiddenParts entry
// a positive hit case (an identifier whose splitIdentifier output
// contains the exact want-sequence) and a near-miss case (a superficially
// similar identifier that must NOT match), per M2-T3's AC.
func TestMatchesForbiddenParts_HitAndNearMiss(t *testing.T) {
	cases := []struct {
		token    string
		hit      string
		nearMiss string
	}{
		{"slack", "SlackClient", "SlacklikeClient"},
		{"socketmode", "SocketMode", "SocketModel"},
		{"channel_id", "ChannelID", "ChannelIdentifier"},
		{"team_id", "TeamID", "TeamIdentifier"},
		{"acp", "ACPHandler", "Acpress"},
		{"host_cli", "HostCLI", "HostClient"},
		{"ipc_name", "IPCName", "IPCNamespace"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.token, func(t *testing.T) {
			token, _, ok := matchesForbiddenParts(splitIdentifier(c.hit))
			if !ok || token != c.token {
				t.Fatalf("matchesForbiddenParts(splitIdentifier(%q)) = (%q, %v), want (%q, true)", c.hit, token, ok, c.token)
			}
			if _, _, ok := matchesForbiddenParts(splitIdentifier(c.nearMiss)); ok {
				t.Fatalf("matchesForbiddenParts(splitIdentifier(%q)) unexpectedly matched; want no match", c.nearMiss)
			}
		})
	}
}

func TestWindowMatches_PluralTolerance(t *testing.T) {
	if !windowMatches([]string{"team", "ids"}, []string{"team", "id"}) {
		t.Fatalf("windowMatches should tolerate a plural 's' suffix on the final element")
	}
}

func TestWindowMatches_PluralTolerance_NearMiss(t *testing.T) {
	if windowMatches([]string{"team", "idses"}, []string{"team", "id"}) {
		t.Fatalf("windowMatches must not double-strip a plural suffix")
	}
}

func TestWindowMatches_ScopedToWindowNotStringEnd(t *testing.T) {
	// 'TeamIDsCache' -> ['team','ids','cache']: the plural-suffix rule
	// must fire scoped to the ['team','ids'] WINDOW, not anchored to the
	// end of the whole parts list.
	parts := splitIdentifier("TeamIDsCache")
	if !windowMatches(parts, []string{"team", "id"}) {
		t.Fatalf("windowMatches(%v, [team id]) = false, want true (window-scoped plural tolerance)", parts)
	}
}
