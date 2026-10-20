package app

// Opt-in live-SDK end-to-end test (070.004-T). It drives the REAL Copilot
// runtime through the production factory [CopilotAgent] (so no SDK import is
// needed here; the depguard rule confines the SDK to internal/agent/copilot),
// the real hub and WebSocket server, and the reference client.
//
// It is skipped unless INTERCOM_LIVE_SDK_TESTS opts in (same default-deny
// semantics as internal/agent/copilot/live_test.go: strconv.ParseBool, deny on
// any parse error). It covers scenario steps 2, 5 and 7 of the primary-scenario
// deliberation against a real runtime:
//
//	step 2  the SDK client starts (Start is the readiness gate)
//	step 5  a prompt runs a turn that raises a permission request, answered
//	        through the reference client
//	step 7  the operator quits; shutdown completes and no CLI child remains
//
// SQ-b (serialised session callbacks) and SQ-c (permission dispatch is
// independent of event dispatch) need the SDK callback seam, which only the
// adapter package may touch; they are re-verified by
// internal/agent/copilot.TestLiveSQbSQcReverification (see
// docs/runbooks/intercom-operator-runbook.md for the commands).
//
// Run it:
//
//	INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/app -run Live -v
//	$env:INTERCOM_LIVE_SDK_TESTS = "1"; go test ./internal/app -run Live -v   # PowerShell
//
// Prerequisites: the Copilot CLI installed and authenticated (or a GitHub token
// in the environment the CLI reads, for example GH_TOKEN or GITHUB_TOKEN).
// Optional: INTERCOM_LIVE_PROMPT replaces the prompt; INTERCOM_LIVE_CLI_PROCESS
// replaces the process-name needle used by the stranded-child check
// (default "copilot"); [copilot].cli_path is taken from INTERCOM_LIVE_CLI_PATH
// when set (an absolute path).

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

const (
	liveSDKTestsEnvVar = "INTERCOM_LIVE_SDK_TESTS"
	// liveWait bounds each live wait (model latency, CLI startup).
	liveWait = 3 * time.Minute
	// livePrompt asks for a harmless read inside the temp workspace.
	livePrompt = "Use your file reading tool to read the file notes.txt in the current directory, then reply with only the word DONE."
)

// liveSDKTestsAllowed reports whether raw opts in: strconv.ParseBool, and deny
// on any parse error or empty value.
func liveSDKTestsAllowed(raw string) bool {
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	return err == nil && v
}

// readInsideWorkspace reports whether the read permission's summary (the
// requested path, absolute or relative to ws) names an existing location that
// lies inside ws after symlinks are resolved. Anything else, including an
// empty or unresolvable summary, is not confirmed inside, so it is denied.
func readInsideWorkspace(ws, summary string) bool {
	summary = strings.TrimSpace(summary)
	if summary == "" || strings.ContainsRune(summary, 0) {
		return false
	}
	root, err := filepath.EvalSymlinks(ws)
	if err != nil {
		return false
	}
	p := summary
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	real, err := filepath.EvalSymlinks(p)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func TestLiveReadApprovalIsConfinedToWorkspace(t *testing.T) {
	ws := t.TempDir()
	inside := filepath.Join(ws, "notes.txt")
	if err := os.WriteFile(inside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	for summary, want := range map[string]bool{
		inside:                                   true,
		"notes.txt":                              true,
		outside:                                  false,
		"../" + filepath.Base(outside):           false,
		"":                                       false,
		"missing.txt":                            false,
		filepath.Join(ws, "..", "..", "x"):       false,
		"notes.txt\x00" + filepath.Base(outside): false,
	} {
		if got := readInsideWorkspace(ws, summary); got != want {
			t.Errorf("readInsideWorkspace(%q) = %v; want %v", summary, got, want)
		}
	}
}

func TestLiveOptInSemantics(t *testing.T) {
	for raw, want := range map[string]bool{"": false, "1": true, "0": false, "false": false, "off": false, "true": true, "garbage": false} {
		if got := liveSDKTestsAllowed(raw); got != want {
			t.Errorf("liveSDKTestsAllowed(%q) = %v; want %v", raw, got, want)
		}
	}
}

// TestLiveCopilotEndToEnd is the real-runtime scenario. See the file comment.
func TestLiveCopilotEndToEnd(t *testing.T) {
	if !liveSDKTestsAllowed(os.Getenv(liveSDKTestsEnvVar)) {
		t.Skipf("%s is not enabled; skipping the live Copilot SDK test (run: %s=1 go test ./internal/app -run Live -v)", liveSDKTestsEnvVar, liveSDKTestsEnvVar)
	}

	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "notes.txt"), []byte("intercom live test notes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tables := "[copilot]\nstartup_timeout_seconds = 120\n"
	if p := os.Getenv("INTERCOM_LIVE_CLI_PATH"); p != "" {
		tables += "cli_path = '" + p + "'\n"
	}
	needle := os.Getenv("INTERCOM_LIVE_CLI_PROCESS")
	if needle == "" {
		needle = "copilot"
	}

	listCtx, listCancel := context.WithTimeout(context.Background(), waitFor)
	baselineTable, err := listProcesses(listCtx)
	listCancel()
	if err != nil {
		t.Fatalf("cannot snapshot the process table, so a stranded CLI child could not be detected: %v", err)
	}
	baseline := pidSet(baselineTable)

	sig := make(chan os.Signal, 2)
	logs := &logBuf{}
	startCtx, cancel := context.WithTimeout(context.Background(), liveWait)
	defer cancel()
	a, err := Start(startCtx, Options{
		ConfigPath: writeConfig(t, ws, "", tables),
		Port:       0,
		PortSet:    true,
		NewAgent:   CopilotAgent,
		Logger:     newTestLogger(logs),
		Signals:    sig,
		TUI:        TUIOptions{Disabled: true},
	})
	if err != nil {
		t.Fatalf("Start against the real runtime: %v", err) // step 2
	}
	shutDown := false
	t.Cleanup(func() {
		if !shutDown {
			stopApp(t, a)
		}
	})

	// The method must be able to see the child, or "none left" proves nothing.
	// The CLI may take a moment to appear in the table after Start returns.
	var children map[int]string
	for deadline := time.Now().Add(waitFor); ; {
		table := mustListProcesses(t)
		if children = newChildren(table, os.Getpid(), baseline, needle); len(children) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("no CLI child process was visible after Start (needle %q): the stranded-child check would be vacuous; set INTERCOM_LIVE_CLI_PROCESS to the CLI's process name", needle)
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Logf("CLI process tree after Start: %v", children)

	// Step 5: a prompt that triggers a permission request, answered remotely.
	approver := dialLive(t, a, hub.RoleApprover)
	approver.waitFor(t, "session.hydrate", func(m wsclient.Message) bool { return m.Envelope != nil && m.Envelope.Kind == hub.KindSessionHydrate })

	prompt := livePrompt
	if p := os.Getenv("INTERCOM_LIVE_PROMPT"); p != "" {
		prompt = p
	}
	pctx, pcancel := context.WithTimeout(context.Background(), waitFor)
	err = approver.c.Prompt(pctx, prompt)
	pcancel()
	if err != nil {
		t.Fatalf("send prompt: %v", err)
	}
	reply := approver.waitFor(t, "prompt reply", func(m wsclient.Message) bool { return m.Reply != nil })
	if !reply.Reply.Accepted() {
		t.Fatalf("prompt refused: %+v", *reply.Reply)
	}

	req := approver.waitFor(t, "permission.requested", func(m wsclient.Message) bool {
		return m.Envelope != nil && m.Envelope.Kind == hub.KindPermissionRequested
	})
	var pr hub.PermissionRequestedPayload
	if err := json.Unmarshal(req.Envelope.Payload, &pr); err != nil {
		t.Fatalf("decode permission request: %v", err)
	}
	// Never approve anything but a read of a path confirmed inside the temp
	// workspace, whatever the model decided to do; deny everything else.
	allow := pr.PermissionKind == "read" && readInsideWorkspace(ws, pr.Summary)
	rctx, rcancel := context.WithTimeout(context.Background(), waitFor)
	err = approver.c.ResolvePermission(rctx, pr.RequestID, allow, "intercom live test")
	rcancel()
	if err != nil {
		t.Fatalf("send resolve_permission: %v", err)
	}
	ans := approver.waitFor(t, "resolve reply", func(m wsclient.Message) bool { return m.Reply != nil })
	if !ans.Reply.Accepted() {
		t.Fatalf("resolve_permission refused: %+v (kind %q)", *ans.Reply, pr.PermissionKind)
	}
	resolved := approver.waitFor(t, "permission.resolved", func(m wsclient.Message) bool {
		return m.Envelope != nil && m.Envelope.Kind == hub.KindPermissionResolved
	})
	var rp hub.PermissionResolvedPayload
	if err := json.Unmarshal(resolved.Envelope.Payload, &rp); err != nil {
		t.Fatalf("decode permission.resolved: %v", err)
	}
	if rp.RequestID != pr.RequestID || rp.Actor != "remote-approver" {
		t.Errorf("permission.resolved = %+v; want the remote approver's answer to %s", rp, pr.RequestID)
	}
	// The turn finishes (a denied request still ends the turn).
	approver.waitFor(t, "agent.idle", func(m wsclient.Message) bool { return m.Envelope != nil && m.Envelope.Kind == hub.KindAgentIdle })

	// Step 7: shut down like an operator pressing Ctrl+C, then prove no CLI child is left.
	sig <- os.Interrupt
	select {
	case <-a.Done():
		shutDown = true
	case <-time.After(liveWait):
		t.Fatal("the process did not finish shutting down")
	}
	if err := a.Err(); err != nil {
		t.Errorf("shutdown reported errors: %v", err)
	}
	for deadline := time.Now().Add(waitFor); ; {
		left := newChildren(mustListProcesses(t), os.Getpid(), baseline, needle)
		if len(left) == 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("CLI child process(es) stranded after shutdown: %v", left)
		}
		time.Sleep(250 * time.Millisecond)
	}

	for _, c := range a.Credentials() {
		if strings.Contains(logs.String(), c.Token.Reveal()) {
			t.Fatal("a credential appears in the logs")
		}
	}
}

func mustListProcesses(t *testing.T) []procEntry {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	table, err := listProcesses(ctx)
	if err != nil {
		t.Fatalf("process listing: %v", err)
	}
	return table
}

// liveSurface is a reference-client connection with live-length waits.
type liveSurface struct {
	c     *wsclient.Client
	queue []wsclient.Message
}

func dialLive(t *testing.T, a *App, role hub.Role) *liveSurface {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: a.WSURL(), Credential: credentialFor(t, a, role)})
	if err != nil {
		t.Fatalf("dial as role %v: %v", role, err)
	}
	t.Cleanup(c.Abort)
	return &liveSurface{c: c}
}

// waitFor returns the first frame satisfying pred, reading for up to liveWait
// in total; other frames are kept so nothing is lost between waits.
func (s *liveSurface) waitFor(t *testing.T, what string, pred func(wsclient.Message) bool) wsclient.Message {
	t.Helper()
	for i, m := range s.queue {
		if pred(m) {
			s.queue = append(s.queue[:i:i], s.queue[i+1:]...)
			return m
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), liveWait)
	defer cancel()
	for {
		m, err := s.c.Next(ctx)
		if err != nil {
			t.Fatalf("waiting for %s: %v", what, err)
		}
		if pred(m) {
			return m
		}
		s.queue = append(s.queue, m)
	}
}
