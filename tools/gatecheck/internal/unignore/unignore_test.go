package unignore

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// failingGit wraps real so that every invocation of the given git
// subcommand fails with a plain error, letting a test drive a chosen
// callee failure through unignore's print sites.
func failingGit(real GitRunner, subcommand string) GitRunner {
	return func(dir string, stdin []byte, args ...string) ([]byte, []byte, error) {
		if len(args) > 0 && args[0] == subcommand {
			return nil, []byte("injected"), errors.New("boom")
		}
		return real(dir, stdin, args...)
	}
}

// denylistRepo returns a committed repository whose root .gitignore ignores
// every Denylist entry, so the only stderr lines a run emits are the
// injected-failure diagnostics.
func denylistRepo(t *testing.T) string {
	t.Helper()
	root := initRepoWithCommit(t, newIsolatedGitRunner(t))
	if err := writeTestGitignore(root, strings.Join(Denylist, "\n")+"\n"); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	return root
}

// assertSingleErrorPrefix requires every non-empty stderr line to start
// with exactly one "::error::" prefix, whatever the callee error text was.
func assertSingleErrorPrefix(t *testing.T, stderr string) {
	t.Helper()
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("stderr is empty, want at least one ::error:: line")
	}
	for _, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
		if !strings.HasPrefix(line, "::error::") {
			t.Errorf("stderr line %q lacks the leading ::error:: prefix", line)
		}
		if strings.HasPrefix(line, "::error::::error::") {
			t.Errorf("stderr line %q has a doubled ::error:: prefix", line)
		}
	}
}

// TestPrintSites_SingleErrorPrefix pins the print-site ownership
// convention: callees may return errors with or without a leading
// "::error::", and the print sites emit exactly one.
func TestPrintSites_SingleErrorPrefix(t *testing.T) {
	// "ls-files" yields a callee error already prefixed with ::error::;
	// "init" yields a plain wrapped error.
	// Distinct base/head contents let both scenario commits succeed so the
	// injected failure is the one actually reached.
	for _, sub := range []string{"ls-files", "init"} {
		t.Run("runScenario/"+sub, func(t *testing.T) {
			git := failingGit(newIsolatedGitRunner(t), sub)
			var stdout, stderr bytes.Buffer
			ok := runScenario(git, t.TempDir(), "s", "a\n", "a\n!b\n", nil, false, nil, &stdout, &stderr)
			if ok {
				t.Fatal("expected the injected failure to fail the scenario")
			}
			assertSingleErrorPrefix(t, stderr.String())
		})
		t.Run("runSelfTest/"+sub, func(t *testing.T) {
			git := failingGit(newIsolatedGitRunner(t), sub)
			var stdout, stderr bytes.Buffer
			if code := runSelfTest(denylistRepo(t), git, &stdout, &stderr); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			assertSingleErrorPrefix(t, stderr.String())
		})
		t.Run("runCheck/"+sub, func(t *testing.T) {
			git := failingGit(newIsolatedGitRunner(t), sub)
			var stdout, stderr bytes.Buffer
			root := denylistRepo(t)
			if code := runCheck(root, "HEAD", "HEAD", git, &stdout, &stderr); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			assertSingleErrorPrefix(t, stderr.String())
		})
	}
}

func TestPyListRepr(t *testing.T) {
	if got := pyListRepr(nil); got != "[]" {
		t.Fatalf("pyListRepr(nil) = %q, want []", got)
	}
	if got := pyListRepr([]string{"a"}); got != "['a']" {
		t.Fatalf("pyListRepr([a]) = %q, want ['a']", got)
	}
	if got := pyListRepr([]string{"a", "b"}); got != "['a', 'b']" {
		t.Fatalf("pyListRepr([a,b]) = %q, want ['a', 'b']", got)
	}
}

func TestDoSelfTestLandingPrecondition_Pass(t *testing.T) {
	git := newIsolatedGitRunner(t)
	repoDir := t.TempDir()
	scratchRoot := t.TempDir()

	var b strings.Builder
	for _, entry := range Denylist {
		b.WriteString(entry)
		b.WriteString("\n")
	}
	if err := writeTestGitignore(repoDir, b.String()); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ok := doSelfTestLandingPrecondition(git, repoDir, scratchRoot, &stdout, &stderr)
	if !ok {
		t.Fatalf("expected PASS, stderr=%q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "PASS landing-precondition") {
		t.Fatalf("stdout = %q, want it to contain PASS landing-precondition", stdout.String())
	}
}

func TestDoSelfTestLandingPrecondition_Fail(t *testing.T) {
	git := newIsolatedGitRunner(t)
	repoDir := t.TempDir()
	scratchRoot := t.TempDir()
	// An empty .gitignore means every denylist entry fails to be ignored.
	if err := writeTestGitignore(repoDir, ""); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	var stdout, stderr bytes.Buffer
	ok := doSelfTestLandingPrecondition(git, repoDir, scratchRoot, &stdout, &stderr)
	if ok {
		t.Fatal("expected FAIL")
	}
	if !strings.Contains(stderr.String(), "FAIL landing-precondition") {
		t.Fatalf("stderr = %q, want it to contain FAIL landing-precondition", stderr.String())
	}
}

func TestRunScenario_PassAndFail(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()

	var stdout, stderr bytes.Buffer
	ok := runScenario(
		git, tmpRoot, "reject-existing-file-unignored",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		[]string{"foo/bar.secret"}, true, nil, &stdout, &stderr,
	)
	if !ok {
		t.Fatalf("expected the scenario's own PASS/FAIL assertion to hold, stderr=%q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "PASS reject-existing-file-unignored: rejected as expected") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	// Assert the OPPOSITE of what actually happens (expect clean when the
	// scenario actually rejects) to exercise the FAIL branch.
	ok = runScenario(
		git, tmpRoot, "expect-clean-but-rejects",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		[]string{"foo/bar.secret"}, false, nil, &stdout, &stderr,
	)
	if ok {
		t.Fatal("expected the mismatched expectation to report FAIL")
	}
	if !strings.Contains(stderr.String(), "FAIL expect-clean-but-rejects: expected clean, got rejection") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestDoSelfTestScenarios(t *testing.T) {
	git := newIsolatedGitRunner(t)
	var stdout, stderr bytes.Buffer
	ok := doSelfTestScenarios(git, &stdout, &stderr)
	if !ok {
		t.Fatalf("expected all three scenarios to pass, stderr=%q", stderr.String())
	}
	for _, name := range []string{
		"reject-existing-file-unignored",
		"accept-nonexistent-negation",
		"reject-tracked-file-unignored-in-same-change",
	} {
		if !strings.Contains(stdout.String(), "PASS "+name) {
			t.Errorf("stdout missing PASS line for scenario %q: %q", name, stdout.String())
		}
	}
}

func TestRun_UnknownFlag_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	var stdout, stderr bytes.Buffer
	code := Run(g.UnknownFlag.Args, t.TempDir(), DefaultGitRunner, &stdout, &stderr)
	if code != g.UnknownFlag.ExitCode {
		t.Fatalf("exit code = %d, want %d", code, g.UnknownFlag.ExitCode)
	}
	if stdout.String() != g.UnknownFlag.Stdout {
		t.Fatalf("stdout = %q, want %q", stdout.String(), g.UnknownFlag.Stdout)
	}
	if !strings.Contains(stderr.String(), "unrecognized argument") {
		t.Fatalf("stderr = %q, want it to mention unrecognized argument", stderr.String())
	}
}

func TestRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--help"}, t.TempDir(), DefaultGitRunner, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(stderr.String(), "Usage:") {
		t.Fatalf("stderr = %q, want a usage message", stderr.String())
	}
}

func TestRun_Check_MissingBaseRef_FailsClosed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(nil, t.TempDir(), DefaultGitRunner, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "--base-ref was not supplied") {
		t.Fatalf("stderr = %q, want the fail-closed --base-ref message", stderr.String())
	}
}

// TestRun_SelfTest_MatchesGolden is the parity anchor for the whole
// unignore port: it runs Run(["--self-test"], ...) against THIS
// repository's real checkout, using DefaultGitRunner exactly as
// production CI would invoke it, and compares stdout/exit code
// byte-for-byte against the golden captured from the original Python
// engine (M3-T1). A real repository checkout is required, and the
// checkout must have no stray untracked files at test time (the
// differential/ls-files checks enumerate the real working tree), so this
// test is skipped if the working tree is not clean.
func TestRun_SelfTest_MatchesGolden(t *testing.T) {
	root := repoRoot(t)
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	if !workingTreeIsClean(t, root) {
		t.Skip("working tree is not clean (untracked files present); skipping exact self-test parity check")
	}

	g := loadGolden(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--self-test"}, root, DefaultGitRunner, &stdout, &stderr)
	if code != g.SelfTest.ExitCode {
		t.Fatalf("exit code = %d, want %d; stderr=%q", code, g.SelfTest.ExitCode, stderr.String())
	}
	if stdout.String() != g.SelfTest.Stdout {
		t.Fatalf("stdout mismatch.\n--- got ---\n%s\n--- want ---\n%s", stdout.String(), g.SelfTest.Stdout)
	}
}
