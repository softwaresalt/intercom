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
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
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

// maxReadSummaryRunes mirrors the adapter's maxSummaryRunes bound in
// internal/agent/copilot/permission.go: sanitizeBound copies at most this many
// runes and appends "…" when the original was longer, so a longer summary or
// one ending in the marker is a lossy copy of the path the SDK will open.
const maxReadSummaryRunes = 2048

// readInsideWorkspace reports whether the read permission's summary (the raw
// requested path, absolute or relative to ws) names an existing non-directory
// that lies inside ws after symlinks are resolved. It fails closed: anything
// it cannot prove inside, including every summary that the adapter may have
// altered or that could mean something else to the SDK runtime, is denied.
//
// The summary is never normalized (no trimming, quote or file: stripping, or
// lexical cleaning of ".." before symlink resolution), because the path
// validated must be exactly the literal path the runtime opens.
func readInsideWorkspace(ws, summary string) bool {
	if rejectSummary(summary) {
		return false
	}
	p := summary
	if runtime.GOOS == "windows" {
		var ok bool
		if p, ok = normalizeWindowsExtended(p); !ok {
			return false
		}
	}
	root, err := pathsafe.NewRoot(ws)
	if err != nil {
		return false
	}
	rel := p
	if filepath.IsAbs(p) {
		if rel, err = relativeToRoot(root, p); err != nil {
			return false
		}
	}
	resolved, err := root.Resolve(rel)
	if err != nil {
		return false
	}
	info, err := os.Stat(resolved)
	return err == nil && !info.IsDir()
}

// rejectSummary reports whether p must be denied before any path resolution:
// it is empty, padded, lossy, truncated, a file: URL, names a device
// namespace or a non-Windows extended prefix, or has a literal ".." segment.
func rejectSummary(p string) bool {
	if p == "" || strings.TrimSpace(p) != p || !utf8.ValidString(p) {
		return true
	}
	if utf8.RuneCountInString(p) > maxReadSummaryRunes || strings.HasSuffix(p, "…") {
		return true
	}
	if strings.IndexFunc(p, isUnsafeSummaryRune) >= 0 {
		return true
	}
	// Split the RAW path before any Clean/Dir/Rel/Resolve can fold it away.
	for _, seg := range strings.FieldsFunc(p, func(r rune) bool { return r == '/' || r == '\\' }) {
		if seg == ".." {
			return true
		}
	}
	return strings.HasPrefix(strings.ToLower(p), "file:") || isDeviceOrExtendedForm(p)
}

// isDeviceOrExtendedForm reports whether p starts with a device (`.`) or
// extended (`?`) namespace prefix: two separators, '.' or '?', then a
// separator, where each separator may be '/' or '\'. The device form is denied
// on every platform. The extended form is denied on non-Windows always, and on
// Windows unless it is exactly the backslash spelling `\\?\`, which
// normalizeWindowsExtended handles; the forward-slash and mixed spellings are
// never normalized, so they are rejected outright.
func isDeviceOrExtendedForm(p string) bool {
	if len(p) < 4 || !isSummarySep(p[0]) || !isSummarySep(p[1]) || !isSummarySep(p[3]) {
		return false
	}
	switch p[2] {
	case '.':
		return true
	case '?':
		return runtime.GOOS != "windows" || p[:4] != `\\?\`
	}
	return false
}

func isSummarySep(b byte) bool { return b == '/' || b == '\\' }

// isUnsafeSummaryRune matches NUL, U+FFFD itself and every control or format
// rune (the runes sanitizeBound replaces with U+FFFD), including \n and \t,
// which sanitizeBound keeps: this check is stricter than sanitizeBound.
func isUnsafeSummaryRune(r rune) bool {
	return r == utf8.RuneError || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
}

// normalizeWindowsExtended rewrites a Windows extended-length path to its
// ordinary form (matching pathsafe.stripUNCPrefix): `\\?\UNC\h\s` becomes
// `\\h\s` (checked first) and `\\?\C:\...` becomes `C:\...`. Any other
// `\\?\` form is rejected. Paths without the prefix are returned unchanged.
// Call it only when runtime.GOOS == "windows".
func normalizeWindowsExtended(p string) (string, bool) {
	if !strings.HasPrefix(p, `\\?\`) {
		return p, true
	}
	if len(p) >= 8 && strings.EqualFold(p[4:8], `UNC\`) {
		return `\\` + p[8:], true
	}
	rest := p[4:]
	if len(rest) >= 3 && rest[1] == ':' && rest[2] == '\\' && isASCIILetter(rest[0]) {
		return rest, true
	}
	return "", false
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// relativeToRoot expresses the absolute path p relative to root, canonicalizing
// its parent directory (so a symlinked parent is judged by its real location)
// while leaving the final component for root.Resolve to check.
func relativeToRoot(root pathsafe.Root, p string) (string, error) {
	dirRoot, err := pathsafe.NewRoot(filepath.Dir(p))
	if err != nil {
		return "", err
	}
	return filepath.Rel(root.Path(), filepath.Join(dirRoot.Path(), filepath.Base(p)))
}

// readFixture is a temp workspace holding notes.txt plus a sibling directory
// outside it holding secret.txt, notes.txt and the sub directory.
type readFixture struct {
	ws, outside string
}

func newReadFixture(t *testing.T) readFixture {
	t.Helper()
	fx := readFixture{ws: t.TempDir(), outside: t.TempDir()}
	writeTestFile(t, filepath.Join(fx.ws, "notes.txt"))
	writeTestFile(t, filepath.Join(fx.outside, "secret.txt"))
	writeTestFile(t, filepath.Join(fx.outside, "notes.txt"))
	if err := os.Mkdir(filepath.Join(fx.outside, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	return fx
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// symlinkOrSkip creates link -> target, skipping only the calling subtest
// when the host does not permit symlinks (Windows without the privilege).
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("symlink creation is not permitted on this host: %v", err)
		}
		t.Fatalf("symlink %q -> %q: %v", link, target, err)
	}
}

func assertRead(t *testing.T, ws, summary string, want bool) {
	t.Helper()
	if got := readInsideWorkspace(ws, summary); got != want {
		t.Errorf("readInsideWorkspace(%.80q) = %v; want %v", summary, got, want)
	}
}

// assertAllDenied checks the plain contained path still resolves (so the
// denials below are not vacuous) and every summary is denied.
func assertAllDenied(t *testing.T, fx readFixture, summaries ...string) {
	t.Helper()
	assertRead(t, fx.ws, "notes.txt", true)
	assertRead(t, fx.ws, filepath.Join(fx.ws, "notes.txt"), true)
	for _, s := range summaries {
		assertRead(t, fx.ws, s, false)
	}
}

// longNestedFile creates a file whose absolute path has exactly runes runes
// (nested 100-rune directories, then a final file name) and returns it.
func longNestedFile(t *testing.T, ws string, runes int) string {
	t.Helper()
	cur := ws
	for {
		remaining := runes - utf8.RuneCountInString(cur) - 1 // minus the separator
		if remaining <= 200 {
			cur = filepath.Join(cur, strings.Repeat("f", remaining))
			break
		}
		cur = filepath.Join(cur, strings.Repeat("d", 100))
	}
	if n := utf8.RuneCountInString(cur); n != runes {
		t.Fatalf("built a %d-rune path; want %d", n, runes)
	}
	if err := os.MkdirAll(filepath.Dir(cur), 0o700); err != nil {
		t.Skipf("host cannot create a %d-rune nested path (long path support): %v", runes, err)
	}
	if err := os.WriteFile(cur, []byte("x"), 0o600); err != nil {
		t.Skipf("host cannot create a %d-rune nested file (long path support): %v", runes, err)
	}
	return cur
}

func TestLiveReadApprovalIsConfinedToWorkspace(t *testing.T) {
	t.Run("basics", func(t *testing.T) {
		fx := newReadFixture(t)
		outside := filepath.Join(fx.outside, "secret.txt")
		if err := os.Mkdir(filepath.Join(fx.ws, "dir"), 0o700); err != nil {
			t.Fatal(err)
		}
		for summary, want := range map[string]bool{
			filepath.Join(fx.ws, "notes.txt"): true,
			"notes.txt":                       true,
			"./notes.txt":                     true,
			outside:                           false,
			"../" + filepath.Base(fx.outside) + "/secret.txt": false,
			"":                                       false,
			"missing.txt":                            false,
			"dir":                                    false,
			fx.ws:                                    false,
			filepath.Join(fx.ws, "..", "..", "x"):    false,
			"notes.txt\x00" + filepath.Base(outside): false,
			"   ":                                    false,
		} {
			assertRead(t, fx.ws, summary, want)
		}
	})

	t.Run("symlink escapes", func(t *testing.T) {
		fx := newReadFixture(t)
		symlinkOrSkip(t, fx.outside, filepath.Join(fx.ws, "outdir"))
		symlinkOrSkip(t, filepath.Join(fx.outside, "secret.txt"), filepath.Join(fx.ws, "leak.txt"))
		assertAllDenied(t, fx,
			"outdir/secret.txt",
			filepath.Join(fx.ws, "outdir", "secret.txt"),
			"leak.txt",
			filepath.Join(fx.ws, "leak.txt"),
		)
	})

	t.Run("extended prefix", func(t *testing.T) {
		fx := newReadFixture(t)
		inside := filepath.Join(fx.ws, "notes.txt")
		// A drive-letter extended path is honoured only on Windows; every other
		// extended or device form is denied everywhere.
		assertRead(t, fx.ws, `\\?\`+inside, runtime.GOOS == "windows")
		assertRead(t, fx.ws, `\\?\`+filepath.Join(fx.outside, "secret.txt"), false)
		assertRead(t, fx.ws, `\\.\`+inside, false)
		assertRead(t, fx.ws, `\\.\C:\notes.txt`, false)
		assertRead(t, fx.ws, `\\?\Volume{00000000-0000-0000-0000-000000000000}\notes.txt`, false)
		assertRead(t, fx.ws, `\\?\GLOBALROOT\notes.txt`, false)
		assertRead(t, fx.ws, `\\?\UNC\`, false)
		assertRead(t, fx.ws, `\\?\`, false)
	})

	// The forward-slash and mixed spellings of the device and extended
	// prefixes are never normalized, so they are denied on every platform
	// (including Windows, where only the backslash `\\?\` form is honoured).
	t.Run("forward-slash device and extended prefix", func(t *testing.T) {
		fx := newReadFixture(t)
		inside := filepath.Join(fx.ws, "notes.txt")
		assertAllDenied(t, fx,
			`//./C:/notes.txt`,
			`//?/C:/notes.txt`,
			`\\?/C:\notes.txt`,
			`/\?\C:\notes.txt`,
			`//./`+filepath.ToSlash(inside),
			`//?/`+filepath.ToSlash(inside),
			`\\./`+inside,
			`//.\`+inside,
			`//?/UNC/outside/share/secret.txt`,
		)
	})

	t.Run("long contained path", func(t *testing.T) {
		for _, runes := range []int{600, maxReadSummaryRunes} {
			fx := newReadFixture(t)
			long := longNestedFile(t, fx.ws, runes)
			if n := utf8.RuneCountInString(long); n < 512 || n > maxReadSummaryRunes {
				t.Fatalf("fixture has %d runes; want 512..%d", n, maxReadSummaryRunes)
			}
			assertRead(t, fx.ws, long, true)
			rel, err := filepath.Rel(fx.ws, long)
			if err != nil {
				t.Fatal(err)
			}
			assertRead(t, fx.ws, rel, true)
		}
	})

	// (a) \\?\UNC\outside\share\secret.txt must not collapse onto the local
	// <ws>/UNC/outside/share/secret.txt lookalike.
	t.Run("a external UNC vs local collision", func(t *testing.T) {
		fx := newReadFixture(t)
		local := filepath.Join(fx.ws, "UNC", "outside", "share", "secret.txt")
		writeTestFile(t, local)
		assertRead(t, fx.ws, local, true)
		assertAllDenied(t, fx, `\\?\UNC\outside\share\secret.txt`, `\\?\unc\outside\share\secret.txt`)
	})

	// (b) file: URLs fail closed, never rewritten to a local path.
	t.Run("b file URL vs local collision", func(t *testing.T) {
		fx := newReadFixture(t)
		writeTestFile(t, filepath.Join(fx.ws, "outside", "share", "secret.txt"))
		assertRead(t, fx.ws, "outside/share/secret.txt", true)
		assertAllDenied(t, fx,
			"file://outside/share/secret.txt",
			"file:///outside/share/secret.txt",
			"FILE://outside/share/secret.txt",
			"file:notes.txt",
			"file:///"+filepath.ToSlash(filepath.Join(fx.ws, "notes.txt")),
		)
	})

	// (c) literal quotes are preserved, never stripped to the notes.txt lookalike.
	t.Run("c quoted path vs local collision", func(t *testing.T) {
		fx := newReadFixture(t)
		assertAllDenied(t, fx, `'notes.txt'`, `"notes.txt"`)
		// Control: a quoted literal that really is a contained file is read as such.
		writeTestFile(t, filepath.Join(fx.ws, "'quoted.txt'"))
		assertRead(t, fx.ws, `'quoted.txt'`, true)
		assertRead(t, fx.ws, `quoted.txt`, false)
	})
	t.Run("c quoted symlink to outside", func(t *testing.T) {
		fx := newReadFixture(t)
		symlinkOrSkip(t, filepath.Join(fx.outside, "secret.txt"), filepath.Join(fx.ws, `'notes.txt'`))
		if runtime.GOOS != "windows" { // '"' is not a legal Windows file name character
			symlinkOrSkip(t, filepath.Join(fx.outside, "secret.txt"), filepath.Join(fx.ws, `"notes.txt"`))
		}
		assertAllDenied(t, fx, `'notes.txt'`, `"notes.txt"`)
	})

	// (d) leading or trailing whitespace is never trimmed onto notes.txt.
	t.Run("d whitespace vs local collision", func(t *testing.T) {
		fx := newReadFixture(t)
		assertAllDenied(t, fx, " notes.txt ", "notes.txt ", " notes.txt", "\tnotes.txt", "notes.txt\n",
			" "+filepath.Join(fx.ws, "notes.txt"), filepath.Join(fx.ws, "notes.txt")+" ")
	})
	t.Run("d whitespace symlink to outside", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows strips trailing spaces from file names, so the outside-pointing padded link cannot exist")
		}
		fx := newReadFixture(t)
		symlinkOrSkip(t, filepath.Join(fx.outside, "secret.txt"), filepath.Join(fx.ws, " notes.txt "))
		assertAllDenied(t, fx, " notes.txt ")
	})

	// (e) a literal ".." after a symlink is never cleaned away lexically.
	t.Run("e dot-dot symlink collision", func(t *testing.T) {
		fx := newReadFixture(t)
		symlinkOrSkip(t, filepath.Join(fx.outside, "sub"), filepath.Join(fx.ws, "link"))
		assertAllDenied(t, fx,
			"link/../notes.txt",
			filepath.Join(fx.ws, "link")+"/../notes.txt",
			`link\..\notes.txt`,
			filepath.Join(fx.ws, "link")+`\..\notes.txt`,
			"./../notes.txt",
		)
	})

	// (f) sanitizeBound output (U+FFFD, control, truncation) is lossy and denied.
	t.Run("f sanitizeBound lossy and truncated", func(t *testing.T) {
		fx := newReadFixture(t)
		writeTestFile(t, filepath.Join(fx.ws, "notes\uFFFD.txt"))
		assertAllDenied(t, fx,
			"notes\uFFFD.txt",
			filepath.Join(fx.ws, "notes\uFFFD.txt"),
			"notes\xff.txt",
			"notes\x1b.txt",
			"notes\n.txt",
			"notes\u202e.txt",
			"notes\u200b.txt",
			"notes.txt\u2026",
			strings.Repeat("a", maxReadSummaryRunes+1),
			strings.Repeat("a", maxReadSummaryRunes)+"\u2026",
		)
	})

	// (g) on POSIX \\?\C:\notes.txt is an ordinary file name, never rewritten
	// to the <ws>/C:\notes.txt lookalike.
	t.Run("g POSIX extended prefix collision", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip(`<ws>/C:\notes.txt cannot exist on Windows; \\?\C:\ is covered by "extended prefix"`)
		}
		fx := newReadFixture(t)
		writeTestFile(t, filepath.Join(fx.ws, `C:\notes.txt`))
		assertRead(t, fx.ws, `C:\notes.txt`, true)
		assertAllDenied(t, fx, `\\?\C:\notes.txt`)
	})
	t.Run("g POSIX extended prefix symlink to outside", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip(`\\?\ file names cannot exist on Windows`)
		}
		fx := newReadFixture(t)
		writeTestFile(t, filepath.Join(fx.ws, `C:\notes.txt`))
		symlinkOrSkip(t, filepath.Join(fx.outside, "secret.txt"), filepath.Join(fx.ws, `\\?\C:\notes.txt`))
		assertAllDenied(t, fx, `\\?\C:\notes.txt`)
	})
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
	// Snapshot the credentials first: shutdown may clear the app's own copy, which
	// would turn the log-leak loop below into a vacuous zero-iteration check.
	issued := a.Credentials()
	if len(issued) == 0 {
		t.Fatal("no credentials were issued before shutdown, so the log-leak check would be vacuous")
	}
	for _, c := range issued {
		// strings.Contains(logs, "") is always true, so an empty token must fail loudly.
		if c.Token.Reveal() == "" {
			t.Fatal("an issued credential has an empty token, so the log-leak check would be meaningless")
		}
	}
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

	for _, c := range issued {
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
