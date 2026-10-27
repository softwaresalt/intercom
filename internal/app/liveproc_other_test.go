//go:build !windows

package app

// POSIX process-table snapshot for the live-SDK test's "no stranded CLI child"
// assertion (see liveproc_test.go): ps -axo pid=,ppid=,args= yields every
// process's pid, parent pid, and full command line.

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// listProcesses returns the process table, minus the listing process itself
// and anything it spawned.
func listProcesses(ctx context.Context) ([]procEntry, error) {
	cmd := exec.CommandContext(ctx, "ps", "-axo", "pid=,ppid=,args=") //nolint:gosec // constant arguments
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start process listing: %w", err)
	}
	lister := cmd.Process.Pid
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("process listing: %w", err)
	}
	table := parseProcessTable(out.String())
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

// parseProcessTable parses `ps -axo pid=,ppid=,args=` output; lines that do
// not start with two integers are skipped.
func parseProcessTable(s string) []procEntry {
	var out []procEntry
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pidS, rest, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		rest = strings.TrimSpace(rest)
		ppidS, rest, _ := strings.Cut(rest, " ")
		rest = strings.TrimSpace(rest)
		pid, err1 := strconv.Atoi(strings.TrimSpace(pidS))
		ppid, err2 := strconv.Atoi(strings.TrimSpace(ppidS))
		if err1 != nil || err2 != nil {
			continue
		}
		out = append(out, procEntry{pid: pid, ppid: ppid, cmd: rest})
	}
	return out
}
