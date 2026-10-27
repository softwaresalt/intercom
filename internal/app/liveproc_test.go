package app

// Process-table helpers for the live-SDK test's "no stranded CLI child"
// assertion (070.004-T), and their own always-on test.
//
// Method (portable, no SDK import and no adapter hook needed): the Copilot SDK
// starts the CLI with exec.Command, so the CLI (and anything it spawns) is a
// descendant of the intercom process, which is the test process here. The
// test snapshots the OS process table (Windows: native Toolhelp32 snapshot
// plus a PEB command-line lookup; Linux/macOS: ps -axo pid=,ppid=,args=)
// before Start and again
// after shutdown, and requires that no process that was not in the baseline
// is still a descendant of the test process. A CLI shim whose parent already
// exited is caught by a second, name-based rule: a new process whose command
// line contains the needle (default "copilot", overridable with
// INTERCOM_LIVE_CLI_PROCESS). Only pid and executable name are ever logged,
// never full command lines.

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type procEntry struct {
	pid, ppid int
	cmd       string
}

// listProcesses(ctx) ([]procEntry, error) is platform-specific; see
// liveproc_windows_test.go and liveproc_other_test.go.

// descendants is every process whose parent chain reaches root (root excluded).
func descendants(table []procEntry, root int) map[int]bool {
	kids := map[int][]int{}
	for _, p := range table {
		kids[p.ppid] = append(kids[p.ppid], p.pid)
	}
	out := map[int]bool{}
	queue := []int{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, k := range kids[cur] {
			if !out[k] && k != root {
				out[k] = true
				queue = append(queue, k)
			}
		}
	}
	return out
}

// pidSet is the set of pids in a table.
func pidSet(table []procEntry) map[int]bool {
	m := make(map[int]bool, len(table))
	for _, p := range table {
		m[p.pid] = true
	}
	return m
}

// newChildren returns pid -> executable name for every process in table that
// is absent from baseline and either descends from root or matches needle
// (needle "" disables the name rule). It never returns a command line.
func newChildren(table []procEntry, root int, baseline map[int]bool, needle string) map[int]string {
	desc := descendants(table, root)
	out := map[int]string{}
	for _, p := range table {
		if baseline[p.pid] || p.pid == root {
			continue
		}
		byName := needle != "" && strings.Contains(strings.ToLower(p.cmd), strings.ToLower(needle))
		if desc[p.pid] || byName {
			out[p.pid] = executableName(p.cmd)
		}
	}
	return out
}

func executableName(cmdline string) string {
	cmdline = strings.TrimSpace(cmdline)
	if strings.HasPrefix(cmdline, `"`) {
		if end := strings.Index(cmdline[1:], `"`); end >= 0 {
			cmdline = cmdline[1 : 1+end]
		}
	} else if i := strings.IndexByte(cmdline, ' '); i >= 0 {
		cmdline = cmdline[:i]
	}
	return filepath.Base(strings.ReplaceAll(cmdline, `\`, "/"))
}

// ---- the always-on test of the method itself ----

const procHelperEnv = "INTERCOM_PROC_HELPER"

// TestProcessHelper is the child the next test spawns: it blocks until its
// stdin closes. Without the env var it is a no-op.
func TestProcessHelper(t *testing.T) {
	if os.Getenv(procHelperEnv) != "1" {
		t.Skip("helper process for TestProcessTreeSeesAndForgetsAChild")
	}
	buf := make([]byte, 1)
	_, _ = os.Stdin.Read(buf)
}

// TestProcessTreeSeesAndForgetsAChild proves the live test's detection method
// works on this platform: a spawned child is seen as a new descendant, and is
// gone from the table after it exits. A method that cannot see a child would
// make "no stranded child" vacuous.
func TestProcessTreeSeesAndForgetsAChild(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	before, err := listProcesses(ctx)
	if err != nil {
		t.Skipf("cannot list processes on this platform: %v", err)
	}
	baseline := pidSet(before)

	child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestProcessHelper$") //nolint:gosec // re-executes the test binary
	child.Env = append(os.Environ(), procHelperEnv+"=1")
	stdin, err := child.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := child.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	reaped := false
	defer func() {
		if !reaped {
			_ = child.Process.Kill()
			_ = child.Wait()
		}
	}()

	during, err := listProcesses(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := newChildren(during, os.Getpid(), baseline, "")[child.Process.Pid]; !ok {
		t.Fatalf("the spawned child (pid %d) was not seen as a new descendant", child.Process.Pid)
	}

	_ = stdin.Close()
	if err := child.Wait(); err != nil {
		t.Fatalf("helper exited with an error: %v", err)
	}
	reaped = true
	deadline := time.Now().Add(waitFor)
	for {
		after, err := listProcesses(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, still := newChildren(after, os.Getpid(), baseline, "")[child.Process.Pid]; !still {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the exited child (pid %d) is still listed", child.Process.Pid)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// ---- regression: a reparented child is found by command line alone ----

const (
	procOrphanEnv      = "INTERCOM_PROC_ORPHAN"
	orphanSentinel     = "copilot-orphan-check"
	orphanPIDPrefix    = "orphan-pid="
	orphanMaxLifetime  = 2 * time.Minute // self-limit so a leaked grandchild cannot linger
	orphanHelperTestRE = "-test.run=^TestProcessOrphanHelper$"
)

// TestProcessOrphanHelper plays both roles of the orphan regression test,
// chosen by procOrphanEnv; without the env var it is a no-op.
//
//   - "launcher": starts the grandchild detached (nil stdio, never waited on),
//     prints its pid, and exits at once, orphaning it.
//   - "grandchild": its executable is the test binary (not the sentinel), but
//     its arguments carry the sentinel; it idles until killed or self-limits.
func TestProcessOrphanHelper(t *testing.T) {
	switch os.Getenv(procOrphanEnv) {
	case "launcher":
		gc := exec.Command(os.Args[0], orphanHelperTestRE, "--", "-intercom-sentinel="+orphanSentinel) //nolint:gosec // re-executes the test binary
		gc.Env = append(os.Environ(), procOrphanEnv+"=grandchild")
		if err := gc.Start(); err != nil {
			t.Fatalf("start grandchild: %v", err)
		}
		// Stdout is the test's pipe: the pid line is the only thing it needs.
		_, _ = os.Stdout.WriteString(orphanPIDPrefix + strconv.Itoa(gc.Process.Pid) + "\n")
		_ = gc.Process.Release()
	case "grandchild":
		time.Sleep(orphanMaxLifetime)
	default:
		t.Skip("helper process for TestProcessTreeFindsOrphanedGrandchildByCommandLine")
	}
}

// TestProcessTreeFindsOrphanedGrandchildByCommandLine proves the needle rule
// finds a surviving process whose parent already exited (so it is no longer a
// descendant of the test process) and whose executable name is not the needle:
// only the full command line carries it. A table that lists executable names
// alone would make the live test's "no stranded CLI child" check vacuous for
// a reparented node.exe ...\copilot\index.js style child.
func TestProcessTreeFindsOrphanedGrandchildByCommandLine(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*waitFor) // launcher + poll deadline below
	defer cancel()
	before, err := listProcesses(ctx)
	if err != nil {
		t.Skipf("cannot list processes on this platform: %v", err)
	}
	baseline := pidSet(before)

	launcher := exec.CommandContext(ctx, os.Args[0], orphanHelperTestRE) //nolint:gosec // re-executes the test binary
	launcher.Env = append(os.Environ(), procOrphanEnv+"=launcher")
	out, err := launcher.Output() // returns once the launcher has exited
	if err != nil {
		t.Fatalf("launcher: %v\n%s", err, out)
	}
	gcPID := 0
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), orphanPIDPrefix); ok {
			gcPID, _ = strconv.Atoi(strings.TrimSpace(v))
		}
	}
	if gcPID <= 0 {
		t.Fatalf("launcher did not report a grandchild pid; output: %q", out)
	}
	t.Cleanup(func() {
		if p, err := os.FindProcess(gcPID); err == nil {
			_ = p.Kill()
			_, _ = p.Wait() // only succeeds where the OS allows waiting on a non-child
		}
	})

	deadline := time.Now().Add(waitFor)
	for {
		table, err := listProcesses(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if desc := descendants(table, os.Getpid()); desc[gcPID] {
			t.Fatalf("grandchild (pid %d) is still a descendant of the test process; it is not orphaned", gcPID)
		}
		if _, byDescent := newChildren(table, os.Getpid(), baseline, "")[gcPID]; byDescent {
			t.Fatalf("grandchild (pid %d) was found without the needle; the regression would be vacuous", gcPID)
		}
		found, ok := newChildren(table, os.Getpid(), baseline, orphanSentinel)[gcPID]
		if ok {
			if strings.Contains(strings.ToLower(found), orphanSentinel) {
				t.Fatalf("grandchild executable name %q contains the needle; it must be matched by arguments", found)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("orphaned grandchild (pid %d) was not found by its command-line needle %q", gcPID, orphanSentinel)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
