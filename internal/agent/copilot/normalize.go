package copilot

import (
	"log/slog"

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

// normalize is the single event normalization table (design §4.3, §5.2): it
// maps one SDK session event to an intercom-owned agent.Event. Every SDK
// event variant not listed here yields an explicit KindUnknown result and is
// logged at warn level; unknown events are never silently dropped.
//
// SessionEventData is a sealed, code-generated union (new variants can appear
// in a patch release without a compile error), hence the mandatory default.
func normalize(ev rpc.SessionEvent, log *slog.Logger) agent.Event {
	switch d := ev.Data.(type) {
	case *rpc.AssistantMessageData:
		return agent.Event{Kind: agent.KindAgentMessage, MessageID: d.MessageID, Text: d.Content}
	case *rpc.AssistantMessageDeltaData:
		return agent.Event{Kind: agent.KindDelta, MessageID: d.MessageID, Text: d.DeltaContent}
	case *rpc.AssistantTurnStartData:
		return agent.Event{Kind: agent.KindTurnStart, TurnID: d.TurnID}
	case *rpc.AssistantTurnEndData:
		return agent.Event{Kind: agent.KindTurnEnd, TurnID: d.TurnID}
	case *rpc.SessionIdleData:
		return agent.Event{Kind: agent.KindIdle, Aborted: d.Aborted != nil && *d.Aborted}
	case *rpc.ToolExecutionStartData:
		return agent.Event{Kind: agent.KindToolStarted, ToolCallID: d.ToolCallID, ToolName: d.ToolName}
	case *rpc.ToolExecutionCompleteData:
		return agent.Event{Kind: agent.KindToolEnded, ToolCallID: d.ToolCallID, Success: d.Success}
	default:
		raw := string(ev.Type())
		if log == nil {
			log = slog.Default()
		}
		log.Warn("unhandled agent SDK event type", "event_type", raw)
		return agent.Event{Kind: agent.KindUnknown, RawType: raw}
	}
}
