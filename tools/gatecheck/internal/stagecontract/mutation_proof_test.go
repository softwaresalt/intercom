package stagecontract

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	proofChildEnv     = "STAGECONTRACT_PROOF_CHILD"
	proofChildTimeout = 5 * time.Minute
	proofWaitDelay    = 10 * time.Second
	policiesRel       = ".github/policies/workflow-policies.md"
	stageAgentRel     = ".github/agents/_stage.agent.md"
	p3GrantPrefix     = "- Create and check out"
	p3GrantToken      = "`chore/stage-{scope-slug}`"
)

// testEvent is one `go test -json` event.
type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// childResult is the observed outcome of one nested go test run.
type childResult struct {
	events   []testEvent
	exitCode int
	stderr   string
}

// TestMutationProof proves the contract-to-gate coupling: reverting the P3
// grant in a COPY makes the wired live-contract step fail naming P3, and
// restoring the copy makes it pass. The real checkout is never modified.
// Cost: one nested `go test` per outcome (two total) on a minimal copy.
func TestMutationProof(t *testing.T) {
	if os.Getenv(proofChildEnv) == "1" {
		t.Skip(proofChildEnv + "=1: recursion guard, the proof does not run inside its own child")
	}
	root := proofRepoRoot(t)
	contractFiles := []string{policiesRel, stageAgentRel}

	shaBefore := hashFiles(t, root, contractFiles)
	statusBefore, gitSkip := gitStatusSnapshot(root, contractFiles)

	copyRoot := t.TempDir()
	copyProofTree(t, root, copyRoot)

	policiesCopy := filepath.Join(copyRoot, filepath.FromSlash(policiesRel))
	original, err := os.ReadFile(policiesCopy)
	if err != nil {
		t.Fatalf("read copied contract file: %v", err)
	}
	mutated := mutateP3Grant(t, original)
	if bytes.Equal(original, mutated) {
		t.Fatal("mutation did not change the copied contract bytes")
	}
	if err := os.WriteFile(policiesCopy, mutated, 0o600); err != nil {
		t.Fatalf("write mutated copy: %v", err)
	}

	bad := runLiveContract(t, copyRoot)
	if bad.exitCode == 0 {
		t.Fatal("mutated copy: the live-contract step passed, want a failure")
	}
	if bad.exitCode != 1 {
		t.Fatalf("mutated copy: exit code %d, want 1 (test failure); stderr: %s", bad.exitCode, bad.stderr)
	}
	if !eventNamed(bad.events, "fail", "TestLiveContract") {
		t.Fatalf("mutated copy: no fail event for TestLiveContract in the event stream")
	}
	if !outputMentions(bad.events, "TestLiveContract", "P3: chore/stage-{scope-slug} create and checkout grant is missing") {
		t.Fatalf("mutated copy: failure output does not carry the P3 missing-grant reason")
	}

	if err := os.WriteFile(policiesCopy, original, 0o600); err != nil {
		t.Fatalf("restore copy: %v", err)
	}
	good := runLiveContract(t, copyRoot)
	if good.exitCode != 0 {
		t.Fatalf("restored copy: exit code %d, want 0; stderr: %s", good.exitCode, good.stderr)
	}
	if !eventNamed(good.events, "pass", "TestLiveContract") {
		t.Fatal("restored copy: no pass event for TestLiveContract")
	}

	t.Run("real-tree-sha256", func(t *testing.T) {
		after := hashFiles(t, root, contractFiles)
		for _, f := range contractFiles {
			if shaBefore[f] != after[f] {
				t.Errorf("real contract file %s changed during the proof", f)
			}
		}
	})
	t.Run("real-tree-git-status", func(t *testing.T) {
		if gitSkip != "" {
			t.Skip(gitSkip)
		}
		statusAfter, skip := gitStatusSnapshot(root, contractFiles)
		if skip != "" {
			t.Fatalf("git became unavailable during the proof: %s", skip)
		}
		if !bytes.Equal(statusBefore, statusAfter) {
			t.Errorf("scoped git status changed during the proof:\nbefore: %q\nafter: %q", statusBefore, statusAfter)
		}
	})
}

// proofRepoRoot resolves the repository root from the package directory.
func proofRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func hashFiles(t *testing.T, root string, rels []string) map[string]string {
	t.Helper()
	sums := make(map[string]string, len(rels))
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		sum := sha256.Sum256(data)
		sums[rel] = hex.EncodeToString(sum[:])
	}
	return sums
}

// gitStatusSnapshot returns the scoped porcelain status, or a skip reason
// when git or the repository metadata is unavailable.
func gitStatusSnapshot(root string, contractFiles []string) ([]byte, string) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, "git executable not found on PATH"
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		return nil, ".git is unavailable in the repository root"
	}
	args := []string{"status", "--porcelain=v1", "-z", "--untracked-files=all", "--"}
	args = append(args, contractFiles...)
	args = append(args, "tools/gatecheck/internal/stagecontract")
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Sprintf("git status failed: %v", err)
	}
	return out, ""
}

// copyProofTree copies exactly go.mod, go.sum, tools/gatecheck/** and the two
// contract files into dst, preserving relative depth.
func copyProofTree(t *testing.T, root, dst string) {
	t.Helper()
	copyOne := func(src string) {
		rel, err := filepath.Rel(root, src)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatalf("source %s escapes the repository root", src)
		}
		target := filepath.Join(dst, rel)
		if r, err := filepath.Rel(dst, target); err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
			t.Fatalf("destination %s escapes the temp root", target)
		}
		info, err := os.Lstat(src)
		if err != nil {
			t.Fatalf("lstat %s: %v", src, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
			t.Fatalf("%s is not a regular file (symlinks and special files are refused)", src)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatalf("mkdir for %s: %v", target, err)
		}
		if err := os.WriteFile(target, data, 0o600); err != nil {
			t.Fatalf("write %s: %v", target, err)
		}
	}
	for _, rel := range []string{"go.mod", "go.sum", policiesRel, stageAgentRel} {
		copyOne(filepath.Join(root, filepath.FromSlash(rel)))
	}
	gatecheck := filepath.Join(root, "tools", "gatecheck")
	err := filepath.WalkDir(gatecheck, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		copyOne(path)
		return nil
	})
	if err != nil {
		t.Fatalf("copy tools/gatecheck: %v", err)
	}
}

// mutateP3Grant removes the chore/stage-{scope-slug} token from the Stage
// create/checkout grant line and requires exactly one such grant line.
func mutateP3Grant(t *testing.T, content []byte) []byte {
	t.Helper()
	lines := strings.Split(string(content), "\n")
	hits := 0
	for i, line := range lines {
		if strings.HasPrefix(line, p3GrantPrefix) && strings.Contains(line, p3GrantToken) {
			lines[i] = strings.ReplaceAll(line, " ("+p3GrantToken+")", "")
			hits++
		}
	}
	if hits != 1 {
		t.Fatalf("want exactly one P3 grant line in the copy, found %d", hits)
	}
	return []byte(strings.Join(lines, "\n"))
}

// childEnv builds the explicit allowlisted environment; os.Environ() is never
// passed through, so CI tokens cannot reach the child.
func childEnv(t *testing.T) []string {
	t.Helper()
	env := map[string]string{}
	for _, k := range []string{"PATH", "HOME", "USERPROFILE", "LocalAppData", "AppData", "SystemRoot", "TEMP", "TMP", "PATHEXT", "ComSpec"} {
		if v, ok := os.LookupEnv(k); ok {
			env[k] = v
		}
	}
	for _, k := range []string{"GOROOT", "GOCACHE", "GOMODCACHE", "GOPATH"} {
		out, err := exec.Command("go", "env", k).Output()
		if err != nil {
			t.Fatalf("go env %s: %v", k, err)
		}
		if v := strings.TrimSpace(string(out)); v != "" {
			env[k] = v
		}
	}
	env["GOFLAGS"] = "-mod=readonly"
	env["GOTOOLCHAIN"] = "local"
	env["GOPROXY"] = "off"
	env["GOWORK"] = "off"
	env[proofChildEnv] = "1"
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

// killTree terminates the child and its descendants.
func killTree(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	if runtime.GOOS == "windows" {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", pid).Run()
	} else {
		_ = exec.Command("pkill", "-KILL", "-P", pid).Run()
	}
	return cmd.Process.Kill()
}

// runLiveContract runs the wired step's go test -json command in dir and
// returns the parsed event stream. A deadline, kill or malformed stream fails
// the proof; a non-zero test exit is returned for the caller to judge.
func runLiveContract(t *testing.T, dir string) childResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), proofChildTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-count=1", "-mod=readonly", "-run", "^TestLiveContract$", "./tools/gatecheck/internal/stagecontract/")
	cmd.Dir = dir
	cmd.Env = childEnv(t)
	cmd.WaitDelay = proofWaitDelay
	cmd.Cancel = func() error { return killTree(cmd) }
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("child go test exceeded the %s deadline or was cancelled: %v", proofChildTimeout, ctx.Err())
	}
	res := childResult{stderr: stderr.String()}
	var exitErr *exec.ExitError
	switch {
	case runErr == nil:
	case errors.As(runErr, &exitErr):
		res.exitCode = exitErr.ExitCode()
		if res.exitCode < 0 {
			t.Fatalf("child go test was killed: %v", runErr)
		}
	default:
		t.Fatalf("child go test did not run: %v", runErr)
	}
	res.events = parseEvents(t, &stdout)
	return res
}

// parseEvents decodes a go test -json stream. Any malformed line, an empty
// stream, or a stream without a terminal package event is a failure.
func parseEvents(t *testing.T, r io.Reader) []testEvent {
	t.Helper()
	var events []testEvent
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var ev testEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil || ev.Action == "" {
			t.Fatalf("malformed go test -json event %q: %v", line, err)
		}
		events = append(events, ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("read go test -json stream: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("go test -json stream is empty")
	}
	terminal := false
	for _, ev := range events {
		if ev.Test == "" && (ev.Action == "pass" || ev.Action == "fail") {
			terminal = true
		}
	}
	if !terminal {
		t.Fatal("go test -json stream is truncated: no terminal package event")
	}
	return events
}

func eventNamed(events []testEvent, action, test string) bool {
	for _, ev := range events {
		if ev.Action == action && ev.Test == test {
			return true
		}
	}
	return false
}

func outputMentions(events []testEvent, test, needle string) bool {
	for _, ev := range events {
		if ev.Test == test && ev.Action == "output" && strings.Contains(ev.Output, needle) {
			return true
		}
	}
	return false
}
