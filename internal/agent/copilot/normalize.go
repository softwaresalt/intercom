package copilot

import (
	"log/slog"

	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

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
