package agent_test

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// fakeAgent is an interface-conformance fake: it proves the Agent interface
// is satisfiable without any SDK type.
type fakeAgent struct{}

func (fakeAgent) Start(context.Context) (agent.SessionHandle, error) {
	return agent.SessionHandle{ID: "fake"}, nil
}
func (fakeAgent) Send(context.Context, string) error { return nil }
func (fakeAgent) Abort(context.Context) error        { return nil }
func (fakeAgent) Stop(context.Context) error         { return nil }

var _ agent.Agent = fakeAgent{}

// TestPackageImportsNoSDK enforces the anti-corruption boundary: the
// intercom-owned domain package must not import the Copilot SDK (design 7.1).
func TestPackageImportsNoSDK(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(".", e.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", e.Name(), err)
		}
		checked++
		for _, imp := range f.Imports {
			if strings.Contains(imp.Path.Value, "github.com/github/copilot-sdk") {
				t.Errorf("%s imports the Copilot SDK (%s); the SDK is confined to internal/agent/copilot", e.Name(), imp.Path.Value)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no Go files inspected")
	}
}

func TestEventKindStringsAreStable(t *testing.T) {
	cases := map[agent.EventKind]string{
		agent.KindAgentMessage: "agent.message",
		agent.KindDelta:        "agent.delta",
		agent.KindTurnStart:    "turn.start",
		agent.KindTurnEnd:      "turn.end",
		agent.KindIdle:         "session.idle",
		agent.KindToolStarted:  "tool.started",
		agent.KindToolEnded:    "tool.ended",
		agent.KindUnknown:      "unknown",
	}
	for k, want := range cases {
		if string(k) != want {
			t.Errorf("kind %q != %q", k, want)
		}
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	errs := []error{agent.ErrStopped, agent.ErrNotRunning, agent.ErrAlreadyStarted}
	for i, a := range errs {
		if a == nil || a.Error() == "" {
			t.Fatalf("sentinel %d is empty", i)
		}
		for j, b := range errs {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinels %d and %d are not distinct", i, j)
			}
		}
	}
}

func TestPermissionKindZeroValueIsUnknown(t *testing.T) {
	var k agent.PermissionKind
	if k == agent.PermissionKindShell || k == agent.PermissionKindWrite || k == agent.PermissionKindRead {
		t.Fatal("zero PermissionKind must not alias a real kind (hub policy denies unknown kinds)")
	}
}

func TestDecisionKindZeroValueIsInvalid(t *testing.T) {
	var d agent.PermissionDecision
	if d.Kind.Valid() {
		t.Fatal("zero-value decision kind must not be valid (fail-closed)")
	}
	for _, k := range []agent.DecisionKind{agent.DecisionAllowOnce, agent.DecisionDeny, agent.DecisionUnavailable} {
		if !k.Valid() {
			t.Errorf("kind %v should be valid", k)
		}
	}
}
