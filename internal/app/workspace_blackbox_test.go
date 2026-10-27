package app_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/app"
	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
)

// This file black-boxes workspace resolution containment through the exported
// app.Start (038.001-T); it never calls pathsafe.Root.Resolve itself.
//
//   - Accepted nested paths exercise the Root.Resolve call inside
//     resolveWorkspace (internal/app/workspace.go), the sole production call
//     site, and verify the canonical path is delivered to the agent.
//   - Refusals verify resolveWorkspace's overall fail-closed contract at the
//     app.Start boundary: pathsafe.NewRoot canonicalization plus the
//     filepath.Rel containment check. resolveWorkspace refuses these before
//     reaching Root.Resolve, so the refusal cases are not claimed to exercise
//     Resolve itself.
//
// Expected canonical values come from pathsafe.NewRoot, the same canonicalizer
// production uses, chosen for portability across platforms (Windows 8.3 short
// names, macOS /var -> /private/var).
//
// Files named *_test.go are outside the write-path gate's scan, so ordinary
// fixture setup (os.MkdirAll, os.WriteFile, os.Symlink) is permitted here.

// blackBoxWait bounds every wait; it expires only on a failing run.
const blackBoxWait = 20 * time.Second

// errPrivilegeNotHeld is Windows ERROR_PRIVILEGE_NOT_HELD (1314): the single
// reason a symlink fixture may skip rather than fail on Windows (011.010-T).
const errPrivilegeNotHeld = syscall.Errno(1314)

// stubAgent is the minimal injected agent: it never starts a CLI child.
type stubAgent struct{}

func (stubAgent) Start(context.Context) (agent.SessionHandle, error) {
	return agent.SessionHandle{ID: "blackbox-session"}, nil
}
func (stubAgent) Send(context.Context, string) error { return nil }
func (stubAgent) Abort(context.Context) error        { return nil }
func (stubAgent) Stop(context.Context) error         { return nil }

// depsRecorder is an app.AgentFactory that counts invocations and records the
// working directory it was handed.
type depsRecorder struct {
	mu      sync.Mutex
	calls   int
	workDir string
}

func (r *depsRecorder) build(d app.AgentDeps) (agent.Agent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	r.workDir = d.WorkingDirectory
	return stubAgent{}, nil
}

func (r *depsRecorder) snapshot() (invoked int, workDir string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls, r.workDir
}

// startWorkspace writes a config for defaultRoot (plus one [[workspace]]
// mapping per mapped root) and runs app.Start with workspace as the flag. A
// successful app is shut down on cleanup.
func startWorkspace(t *testing.T, defaultRoot string, mapped []string, workspace string) (*app.App, *depsRecorder, error) {
	t.Helper()
	body := fmt.Sprintf("default_workspace_root = '%s'\n", defaultRoot)
	for i, m := range mapped {
		body += fmt.Sprintf("\n[[workspace]]\nworkspace_id = \"w%d\"\npath = '%s'\n", i, m)
	}
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(cfgPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	rec := &depsRecorder{}
	ctx, cancel := context.WithTimeout(context.Background(), blackBoxWait)
	defer cancel()
	a, err := app.Start(ctx, app.Options{
		ConfigPath: cfgPath,
		Workspace:  workspace,
		PortSet:    true, // port 0: an ephemeral loopback port
		NewAgent:   rec.build,
		TUI:        app.TUIOptions{Disabled: true},
	})
	if err != nil {
		return nil, rec, err
	}
	t.Cleanup(func() {
		a.Shutdown("blackbox cleanup")
		select {
		case <-a.Done():
		case <-time.After(blackBoxWait):
			t.Error("the app did not finish shutting down")
		}
	})
	return a, rec, nil
}

// canonicalPath is the pathsafe-canonical form of an existing directory
// (symlinks and Windows 8.3 short names resolved).
func canonicalPath(t *testing.T, dir string) string {
	t.Helper()
	r, err := pathsafe.NewRoot(dir)
	if err != nil {
		t.Fatalf("pathsafe.NewRoot(%q): %v", dir, err)
	}
	return r.Path()
}

// mkdirAll creates dir and any parents.
func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
}

// requireDelivered asserts the canonical path reached both a.Workspace() and
// AgentDeps.WorkingDirectory, and the agent was built exactly once.
func requireDelivered(t *testing.T, a *app.App, rec *depsRecorder, want string) {
	t.Helper()
	if got := a.Workspace(); got != want {
		t.Errorf("Workspace() = %q; want canonical %q", got, want)
	}
	invoked, workDir := rec.snapshot()
	if invoked != 1 {
		t.Fatalf("agent factory invoked %d times; want 1", invoked)
	}
	if workDir != want {
		t.Errorf("AgentDeps.WorkingDirectory = %q; want canonical %q", workDir, want)
	}
}

// requireRefused asserts app.Start failed closed with KindPathViolation before
// the agent factory ran.
func requireRefused(t *testing.T, a *app.App, rec *depsRecorder, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("app.Start accepted the workspace (Workspace() = %q); want a path violation", a.Workspace())
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind() != apperr.KindPathViolation {
		t.Errorf("error %v is not an apperr.KindPathViolation", err)
	}
	if invoked, _ := rec.snapshot(); invoked != 0 {
		t.Errorf("agent factory invoked %d times; want 0 (refusal must precede NewAgent)", invoked)
	}
}

// symlinkOrSkip creates link -> target. Only the enumerated Windows privilege
// failure (ERROR_PRIVILEGE_NOT_HELD, 1314, on runtime.GOOS == "windows")
// skips; every other failure on every platform fails the test, mirroring
// internal/pathsafe's requireSymlinkOrFailClosed (011.010-T).
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	err := os.Symlink(target, link)
	if err == nil {
		return
	}
	if runtime.GOOS == "windows" && errors.Is(err, errPrivilegeNotHeld) {
		t.Skipf("skipping: ERROR_PRIVILEGE_NOT_HELD -- enable Developer Mode or run elevated: %v", err)
	}
	t.Fatalf("symlink fixture failed for a reason other than the Windows privilege case: %v", err)
}

// TestStartBlackBoxWorkspaceResolveContainment verifies, through app.Start,
// that accepted nested directories under default_workspace_root exercise the
// Root.Resolve call in resolveWorkspace and reach the agent in canonical form.
func TestStartBlackBoxWorkspaceResolveContainment(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "proj", "inner")
	mkdirAll(t, nested)

	// The detour flag is built by concatenation, not filepath.Join, because
	// Join would clean the ".." away. NewRoot's Abs lexically cleans it, but
	// the literal ".." segment must remain in the input handed to app.Start.
	sep := string(filepath.Separator)
	detour := root + sep + "proj" + sep + ".." + sep + "proj" + sep + "inner"

	cases := []struct {
		name      string
		workspace string
	}{
		{"nested directory", nested},
		{"nested directory with an in-root dot-dot detour", detour},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, rec, err := startWorkspace(t, root, nil, tc.workspace)
			if err != nil {
				t.Fatalf("app.Start: %v", err)
			}
			requireDelivered(t, a, rec, canonicalPath(t, nested))
		})
	}
}

// TestStartBlackBoxWorkspaceMappedRootAndRefusals verifies, through app.Start,
// that a nested directory under a [[workspace]].path mapping is accepted (and
// so exercises the Root.Resolve call in resolveWorkspace), and that each escape
// shape fails closed under both the default root and the mapped root. The
// refusals check resolveWorkspace's overall contract (NewRoot canonicalization
// plus the containment check); they are refused before Root.Resolve is reached
// and are not claimed to exercise it.
func TestStartBlackBoxWorkspaceMappedRootAndRefusals(t *testing.T) {
	defaultRoot, mappedRoot := t.TempDir(), t.TempDir()
	mapped := []string{mappedRoot}

	t.Run("nested directory under a mapped root", func(t *testing.T) {
		nested := filepath.Join(mappedRoot, "svc", "api")
		mkdirAll(t, nested)
		a, rec, err := startWorkspace(t, defaultRoot, mapped, nested)
		if err != nil {
			t.Fatalf("app.Start: %v", err)
		}
		requireDelivered(t, a, rec, canonicalPath(t, nested))
	})

	outside := t.TempDir() // outside every authorized root
	for _, base := range []struct{ name, dir string }{
		{"default root", defaultRoot},
		{"mapped root", mappedRoot},
	} {
		escapes := []struct {
			name  string
			build func(t *testing.T) string
		}{
			{"symlink pointing outside every root", func(t *testing.T) string {
				link := filepath.Join(base.dir, "escape")
				symlinkOrSkip(t, outside, link)
				return link
			}},
			{"dot-dot traversal out of the root", func(t *testing.T) string {
				rel, err := filepath.Rel(base.dir, outside)
				if err != nil {
					t.Fatal(err)
				}
				// Concatenated, not filepath.Join, which would clean the ".." away.
				return base.dir + string(filepath.Separator) + rel
			}},
			{"prefix-sibling directory", func(t *testing.T) string {
				sib := base.dir + "-evil"
				// A sibling of a t.TempDir() directory lives inside the test's
				// temp parent, which the testing package removes on cleanup.
				mkdirAll(t, sib)
				return sib
			}},
			{"nonexistent child", func(t *testing.T) string {
				return filepath.Join(base.dir, "absent", "child")
			}},
			{"regular file", func(t *testing.T) string {
				file := filepath.Join(base.dir, "file.txt")
				if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
					t.Fatal(err)
				}
				return file
			}},
		}
		for _, esc := range escapes {
			t.Run(base.name+"/"+esc.name, func(t *testing.T) {
				workspace := esc.build(t)
				a, rec, err := startWorkspace(t, defaultRoot, mapped, workspace)
				requireRefused(t, a, rec, err)
			})
		}
	}
}
