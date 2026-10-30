package copilot

import (
	"log/slog"
	"sync"
	"unicode/utf8"

	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// informationalTypes are SDK event types that carry telemetry or lifecycle
// bookkeeping only. They never change operator-visible state, so they are
// dropped at debug level instead of surfacing as unknown_event errors.
// Anything not listed (including model.call_failure) stays visible.
var informationalTypes = map[string]struct{}{
	"assistant.idle":                    {},
	"assistant.intent":                  {},
	"assistant.message_start":           {},
	"assistant.streaming_delta":         {},
	"assistant.tool_call_delta":         {},
	"assistant.usage":                   {},
	"capabilities.changed":              {},
	"commands.changed":                  {},
	"hook.end":                          {},
	"hook.progress":                     {},
	"hook.start":                        {},
	"mcp.prompts.list_changed":          {},
	"mcp.resources.list_changed":        {},
	"mcp.tools.list_changed":            {},
	"model.call_final_result":           {},
	"model.call_finished":               {},
	"model.call_start":                  {},
	"model.call_started":                {},
	"model.captured_assignment_context": {},
	"model.message":                     {},
	"model.messages_snapshot":           {},
	"model.model_call_started":          {},
	"model.model_call_success":          {},
	"model.response":                    {},
	"model.tool_execution":              {},
	"model.turn_ended":                  {},
	"model.turn_started":                {},
	"pending_messages.modified":         {},
	"permission.completed":              {},
	"permission.requested":              {},
	"sandbox.decision":                  {},
	"session.background_tasks_changed":  {},
	"session.compaction_complete":       {},
	"session.compaction_start":          {},
	"session.shutdown":                  {},
	"session.context_changed":           {},
	"session.custom_agents_updated":     {},
	"session.extensions_loaded":         {},
	"session.managed_settings_enforced": {},
	"session.managed_settings_resolved": {},
	"session.mcp_server_status_changed": {},
	"session.mcp_server_tools_loaded":   {},
	"session.mcp_servers_loaded":        {},
	"session.mode_changed":              {},
	"session.model_change":              {},
	"session.permissions_changed":       {},
	"session.plan_changed":              {},
	"session.remote_steerable_changed":  {},
	"session.session_limits_changed":    {},
	"session.skills_loaded":             {},
	"session.start":                     {},
	"session.title_changed":             {},
	"session.todos_changed":             {},
	"session.tools_updated":             {},
	"session.usage_checkpoint":          {},
	"session.usage_info":                {},
	"session.workspace_file_changed":    {},
	"system.message":                    {},
	"tool.execution_partial_result":     {},
	"tool.execution_progress":           {},
	"tool.shell_output":                 {},
	"user.message":                      {},
}

// isInformational reports whether raw names a purely informational SDK event.
func isInformational(raw string) bool {
	_, ok := informationalTypes[raw]
	return ok
}

// Event size limits (design §4.3). MaxEventTextBytes bounds Event.Text and
// MaxEventIdentBytes bounds every identifier-like Event string.
const (
	MaxEventTextBytes  = 16 << 10
	MaxEventIdentBytes = 128
	// maxWarnedEventTypes caps the unmappedLogGate memory.
	maxWarnedEventTypes = 256
)

// boundUTF8 returns s limited to at most maxBytes bytes without splitting a
// rune, and reports whether it was shortened. For valid UTF-8 it inspects at
// most the 3 bytes before the cut (a rune is at most 4 bytes); invalid input
// is cut at maxBytes after the same bounded look-back, never scanned whole.
func boundUTF8(s string, maxBytes int) (string, bool) {
	if maxBytes < 0 {
		maxBytes = 0
	}
	if len(s) <= maxBytes {
		return s, false
	}
	cut := maxBytes
	for back := 0; back < utf8.UTFMax-1 && cut > 0 && !utf8.RuneStart(s[cut]); back++ {
		cut--
	}
	return s[:cut], true
}

// unmappedLogGate rate-limits unmapped-event Warn logs to the first
// occurrence of each event type. The zero value is ready to use; a nil gate
// never rate-limits. The set is capped at maxWarnedEventTypes so a hostile
// stream of distinct types cannot grow memory; types beyond the cap log at
// Debug only.
type unmappedLogGate struct {
	mu     sync.Mutex
	warned map[string]struct{}
}

// first reports whether raw has not been seen before and records it. It
// returns true for every call on a nil gate and false once the cap is full.
func (g *unmappedLogGate) first(raw string) bool {
	if g == nil {
		return true
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.warned == nil {
		g.warned = make(map[string]struct{})
	}
	if _, seen := g.warned[raw]; seen {
		return false
	}
	if len(g.warned) >= maxWarnedEventTypes {
		return false
	}
	g.warned[raw] = struct{}{}
	return true
}

// boundEvent returns e with the event size limits applied, setting Truncated
// when anything was shortened.
func boundEvent(e agent.Event) agent.Event {
	var t bool
	idents := []*string{&e.MessageID, &e.TurnID, &e.ToolCallID, &e.ToolName, &e.RawType}
	for _, f := range idents {
		var c bool
		*f, c = boundUTF8(*f, MaxEventIdentBytes)
		t = t || c
	}
	var c bool
	e.Text, c = boundUTF8(e.Text, MaxEventTextBytes)
	e.Truncated = e.Truncated || t || c
	return e
}

// toolEndedText extracts the operator-visible result or error text of a
// completed tool. An error message wins over result content.
func toolEndedText(d *rpc.ToolExecutionCompleteData) string {
	if d.Error != nil && d.Error.Message != "" {
		return d.Error.Message
	}
	if d.Result != nil {
		if d.Result.Content != "" {
			return d.Result.Content
		}
		return deref(d.Result.DetailedContent)
	}
	return ""
}

// normalize is normalizeWithGate without unmapped-event rate limiting: every
// unmapped event logs at warn level.
func normalize(ev rpc.SessionEvent, log *slog.Logger) agent.Event {
	return normalizeWithGate(ev, log, nil)
}

// normalizeWithGate is the single event normalization table (design §4.3,
// §5.2): it maps one SDK session event to an intercom-owned agent.Event with
// every string bounded rune-safely (MaxEventTextBytes, MaxEventIdentBytes).
// Every SDK event variant not listed here yields an explicit KindUnknown
// result; the first occurrence of each unmapped type is logged at warn level
// and repeats at debug level (g may be nil to warn every time). Unknown
// events are never silently dropped.
//
// SessionEventData is a sealed, code-generated union (new variants can appear
// in a patch release without a compile error), hence the mandatory default.
func normalizeWithGate(ev rpc.SessionEvent, log *slog.Logger, g *unmappedLogGate) agent.Event {
	switch d := ev.Data.(type) {
	case *rpc.AssistantMessageData:
		return boundEvent(agent.Event{Kind: agent.KindAgentMessage, MessageID: d.MessageID, Text: d.Content})
	case *rpc.AssistantMessageDeltaData:
		return boundEvent(agent.Event{Kind: agent.KindDelta, MessageID: d.MessageID, Text: d.DeltaContent})
	case *rpc.AssistantTurnStartData:
		return boundEvent(agent.Event{Kind: agent.KindTurnStart, TurnID: d.TurnID})
	case *rpc.AssistantTurnEndData:
		return boundEvent(agent.Event{Kind: agent.KindTurnEnd, TurnID: d.TurnID})
	case *rpc.AssistantReasoningData:
		return boundEvent(agent.Event{Kind: agent.KindReasoning, MessageID: d.ReasoningID, Text: d.Content})
	case *rpc.AssistantReasoningDeltaData:
		return boundEvent(agent.Event{Kind: agent.KindReasoningDelta, MessageID: d.ReasoningID, Text: d.DeltaContent})
	case *rpc.SessionIdleData:
		return agent.Event{Kind: agent.KindIdle, Aborted: d.Aborted != nil && *d.Aborted}
	case *rpc.ToolExecutionStartData:
		return boundEvent(agent.Event{Kind: agent.KindToolStarted, ToolCallID: d.ToolCallID, ToolName: d.ToolName})
	case *rpc.ToolExecutionCompleteData:
		return boundEvent(agent.Event{Kind: agent.KindToolEnded, ToolCallID: d.ToolCallID, Success: d.Success, Text: toolEndedText(d)})
	default:
		raw, _ := boundUTF8(sanitizeIdent(string(ev.Type()), MaxEventIdentBytes), MaxEventIdentBytes)
		if log == nil {
			log = slog.Default()
		}
		if g.first(raw) {
			log.Warn("unhandled agent SDK event type", "event_type", raw)
		} else {
			log.Debug("unhandled agent SDK event type (repeat)", "event_type", raw)
		}
		return agent.Event{Kind: agent.KindUnknown, RawType: raw}
	}
}
