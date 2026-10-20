package app

// Process-table helpers for the live-SDK test's "no stranded CLI child"
// assertion (070.004-T), and their own always-on test.
//
// Method (portable, no SDK import and no adapter hook needed): the Copilot SDK
// starts the CLI with exec.Command, so the CLI (and anything it spawns) is a
// descendant of the intercom process, which is the test process here. The
// test snapshots the OS process table (Windows: Win32_Process through
// PowerShell; Linux/macOS: ps -axo pid=,ppid=,args=) before Start and again
// after shutdown, and requires that no process that was not in the baseline
// is still a descendant of the test process. A CLI shim whose parent already
// exited is caught by a second, name-based rule: a new process whose command
// line contains the needle (default "copilot", overridable with
// INTERCOM_LIVE_CLI_PROCESS). Only pid and executable name are ever logged,
// never full command lines.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type procEntry struct {
	pid, ppid int
	cmd       string
}

// listProcesses returns the process table, minus the listing process itself
// and anything it spawned.
func listProcesses(ctx context.Context) ([]procEntry, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", //nolint:gosec // constant arguments
			`Get-CimInstance Win32_Process | ForEach-Object { '{0}|{1}|{2}' -f $_.ProcessId, $_.ParentProcessId, $_.CommandLine }`)
	} else {
		cmd = exec.CommandContext(ctx, "ps", "-axo", "pid=,ppid=,args=") //nolint:gosec // constant arguments
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start process listing: %w", err)
	}
	lister := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("process listing: %w", err)
	}
	table := parseProcessTable(out.String(), runtime.GOOS == "windows")
	own := descendants(table, lister)
	own[lister] = true
	kept := table[:0:0]
	for _, p := range table {
		if !own[p.pid] {
			kept = append(kept, p)
		}
	}
	return kept, nil
}

func parseProcessTable(s string, pipeSeparated bool) []procEntry {
	var out []procEntry
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var pidS, ppidS, rest string
		if pipeSeparated {
			parts := strings.SplitN(line, "|", 3)
			if len(parts) < 2 {
				continue
			}
			pidS, ppidS = parts[0], parts[1]
			if len(parts) == 3 {
				rest = parts[2]
			}
		} else {
			var ok bool
			pidS, rest, ok = strings.Cut(line, " ")
			if !ok {
				continue
			}
			rest = strings.TrimSpace(rest)
			ppidS, rest, _ = strings.Cut(rest, " ")
			rest = strings.TrimSpace(rest)
		}
		pid, err1 := strconv.Atoi(strings.TrimSpace(pidS))
		ppid, err2 := strconv.Atoi(strings.TrimSpace(ppidS))
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, procEntry{pid: pid, ppid: ppid, cmd: rest})
	}
	return out
}

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
