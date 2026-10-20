package app

import (
	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/agent/copilot"
)

// CopilotAgent is the production [AgentFactory]: the Copilot SDK adapter
// (design §7.1), wired to the hub through d. cmd/intercom passes it; tests and
// the end-to-end scenario pass a fake agent instead.
func CopilotAgent(d AgentDeps) (agent.Agent, error) {
	return copilot.New(copilot.Config{
		CLIPath:          d.CLIPath,
		WorkingDirectory: d.WorkingDirectory,
		StartupTimeout:   d.StartupTimeout,
		Ingress:          d.Ingress,
		Asks:             d.Asks,
		HubDone:          d.HubDone,
		Logger:           d.Logger,
	})
}
