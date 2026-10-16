package integration

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// startScriptWaitDelay bounds how long Wait may block on I/O pipes (and on a
// child that ignores cancellation) after the context is done or the child
// exits. start.ps1's engram index worker can outlive the launched process
// while holding the inherited stdout pipe; without a WaitDelay,
// CombinedOutput blocks until that orphan exits (observed 140-206 s).
const startScriptWaitDelay = 5 * time.Second

// startScriptCombinedOutputBounded runs cmd.CombinedOutput with
// cmd.WaitDelay = startScriptWaitDelay.
//
// stdlib semantics, verified against go1.26.9 (GOROOT
// .../toolchain@v0.0.1-go1.26.9.windows-amd64/src/os/exec/exec.go):
//   - WaitDelay doc, lines 278-303: the timer starts when the Context is
//     done or Wait sees the child exit; after it elapses the child is killed
//     and open I/O pipes are closed. Zero (the default) reads pipes to EOF,
//     which may not occur until orphaned descendants exit (lines 300-302).
//   - CommandContext sets Cancel to Process.Kill and leaves WaitDelay unset
//     (lines 481-492); watchCtx calls Cancel on ctx.Done (lines 800-820) and
//     starts the WaitDelay timer (line 826).
//   - ErrWaitDelay is reported only if the process otherwise succeeded and no
//     Cancel occurred (lines 296-298, 871-873, 1002). A context-killed child
//     therefore yields its *exec.ExitError (the Cancel error path, line 809,
//     takes precedence), so callers keep the
//     errors.Is(ctx.Err(), context.DeadlineExceeded) branch first.
//
// WaitDelay is containment only: it does not kill orphaned descendants.
func startScriptCombinedOutputBounded(cmd *exec.Cmd) ([]byte, error) {
	cmd.WaitDelay = startScriptWaitDelay
	return cmd.CombinedOutput()
}

const (
	helperModeEnv = "STARTSCRIPT_HERMETIC_HELPER"
	helperPIDEnv  = "STARTSCRIPT_HERMETIC_PIDFILE"
)

// TestStartScriptHermeticHelperProcess is not a real test: it is the
// re-executed helper for TestStartScriptRunnerBoundedByWaitDelay and returns
// immediately unless helperModeEnv is set.
func TestStartScriptHermeticHelperProcess(t *testing.T) {
	switch os.Getenv(helperModeEnv) {
	case "":
		return
	case "sleep":
		time.Sleep(20 * time.Second)
	case "parent":
		self, err := os.Executable()
		if err != nil {
			t.Fatalf("os.Executable: %v", err)
		}
		gc := exec.Command(self, "-test.run=^TestStartScriptHermeticHelperProcess$")
		gc.Env = append(os.Environ(), helperModeEnv+"=sleep")
		// Inherit the pipe so the grandchild genuinely holds it open.
		gc.Stdout, gc.Stderr = os.Stdout, os.Stderr
		if err := gc.Start(); err != nil {
			t.Fatalf("starting grandchild: %v", err)
		}
		pidFile := os.Getenv(helperPIDEnv)
		tmp := pidFile + ".tmp"
		if err := os.WriteFile(tmp, []byte(strconv.Itoa(gc.Process.Pid)), 0o600); err != nil {
			t.Fatalf("writing pid file: %v", err)
		}
		if err := os.Rename(tmp, pidFile); err != nil {
			t.Fatalf("publishing pid file: %v", err)
		}
		time.Sleep(20 * time.Second)
	}
}

// G3: a cancelled command whose grandchild holds the inherited stdout pipe
// returns within 10 s of cancellation instead of blocking for the
// grandchild's ~20 s lifetime. Needs no pwsh, so it runs on Linux CI.
func TestStartScriptRunnerBoundedByWaitDelay(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	pidFile := filepath.Join(t.TempDir(), "grandchild.pid")
	// Registered after t.TempDir, so it runs before the dir is removed.
	t.Cleanup(func() {
		raw, err := os.ReadFile(pidFile)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(string(raw))
		if err != nil {
			return
		}
		if proc, err := os.FindProcess(pid); err == nil {
			if kerr := proc.Kill(); kerr != nil && !errors.Is(kerr, os.ErrProcessDone) {
				t.Logf("killing grandchild %d: %v (likely already exited)", pid, kerr)
			}
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, self, "-test.run=^TestStartScriptHermeticHelperProcess$")
	cmd.Env = append(os.Environ(), helperModeEnv+"=parent", helperPIDEnv+"="+pidFile)

	type result struct {
		out []byte
		err error
	}
	done := make(chan result, 1)
	go func() {
		out, err := startScriptCombinedOutputBounded(cmd)
		done <- result{out, err}
	}()

	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(pidFile); err == nil {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("grandchild pid file never appeared; guard would be vacuous")
		}
		time.Sleep(50 * time.Millisecond)
	}

	cancelledAt := time.Now()
	cancel()
	select {
	case res := <-done:
		if elapsed := time.Since(cancelledAt); elapsed > 10*time.Second {
			t.Fatalf("runner returned %s after cancel, want <= 10s", elapsed)
		}
		if res.err == nil {
			t.Fatalf("expected an error from a cancelled command, got nil")
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("runner still blocked 10s after cancel: pipe-holding grandchild not contained")
	}
}
