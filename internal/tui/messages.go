package tui

import (
	"encoding/json"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// FrameMsg carries one hub envelope to the model.
type FrameMsg struct{ Frame hub.Frame }

// ResyncMsg notes that the feed overflowed and forced a re-hydrate. It is
// informational: the hydrate that follows is what repairs the model.
type ResyncMsg struct{}

// DisconnectedMsg reports that the hub connection ended for good.
type DisconnectedMsg struct{ Err error }

// ActionResultMsg is the outcome of an action Cmd that went through ingress.
type ActionResultMsg struct {
	Action hub.OperatorAction
	Result hub.ActionResult
	Err    error
}

// entry is one transcript item, mirroring the hub fold's projection. All text
// fields have been through [Sanitize] before they are stored.
type entry struct {
	Kind      string `json:"kind"` // message | tool | error
	ID        string `json:"id"`
	Text      string `json:"text"`
	Tool      string `json:"tool_name"`
	Status    string `json:"status"`
	Truncated bool   `json:"truncated"`
}

// Entry kinds and statuses, as produced by the hub fold.
const (
	kindMessage = "message"
	kindTool    = "tool"
	kindError   = "error"

	statusStreaming = "streaming"
	statusComplete  = "complete"
	statusRunning   = "running"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
)

// hydrateBody is the session.hydrate payload (design §5.3).
type hydrateBody struct {
	Entries            []entry                          `json:"entries"`
	TurnActive         bool                             `json:"turn_active"`
	Truncated          bool                             `json:"truncated"`
	EvictedCount       int                              `json:"evicted_count"`
	PendingPermissions []hub.PermissionRequestedPayload `json:"pending_permissions"`
}

// parseTS parses an envelope timestamp; the zero time on failure.
func parseTS(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// decodeEnvelope decodes one frame.
func decodeEnvelope(f hub.Frame) (hub.Envelope, error) {
	var env hub.Envelope
	err := json.Unmarshal(f.Bytes(), &env)
	return env, err
}
