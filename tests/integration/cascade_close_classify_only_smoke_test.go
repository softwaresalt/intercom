package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	cascadeSmokeShipment  = "001-S"
	cascadeSmokeFeature   = "001-F"
	cascadeSmokeTask      = "001.001-T"
	cascadeSmokeSHA       = "0123456789abcdef0123456789abcdef01234567"
	cascadeSmokeEvidence  = "001-S-001-F-close-evidence.json"
	cascadeSmokeTimeout   = 2 * time.Minute
	cascadeSmokeWaitDelay = 10 * time.Second
)

// cascadeSmokeRegistry is the backlog registry autoharness reads. Removing
// cli.binary makes the tool fail closed; the other keys match the verified
// fixture.
const cascadeSmokeRegistry = "schema_version: 1.0.0\n" +
	"tool_name: backlogit\n" +
	"tool_type: both\n" +
	"directory: .backlog\n" +
	"cli:\n" +
	"  binary: backlogit\n"

// cascadeSmokeDocumentedCodes are the classify-only exit codes documented by
// `autoharness shipment cascade-close --help`.
var cascadeSmokeDocumentedCodes = map[int]bool{0: true, 2: true, 3: true, 4: true}

// cascadeSmokeVerdicts are the classifier_verdict values the contract uses.
var cascadeSmokeVerdicts = map[string]bool{"CASCADE": true, "SAFE_CLOSE": true}

// cascadeSmokeOutcome records one child process run. A timeout is kept apart
// from exit codes so it can never be mistaken for a documented code.
type cascadeSmokeOutcome struct {
	stdout   []byte
	stderr   []byte
	code     int
	timedOut bool
	runErr   error
}

// cascadeSmokeReport is the subset of the --json object this test checks.
type cascadeSmokeReport struct {
	Mode              string   `json:"mode"`
	ExitCode          int      `json:"exit_code"`
	EvidencePath      string   `json:"evidence_path"`
	ClassifierVerdict string   `json:"classifier_verdict"`
	MutationPossible  string   `json:"mutation_possible"`
	Failures          []string `json:"failures"`
}

// cascadeSmokeEvidenceRecord is the subset of the pre_close record checked here.
type cascadeSmokeEvidenceRecord struct {
	ShipmentID string `json:"shipment_id"`
	FeatureID  string `json:"feature_id"`
	Phase      string `json:"phase"`
}

// TestCascadeCloseClassifyOnlySmoke runs the real autoharness cascade-close in
// its read-only classify-only mode against a disposable backlogit fixture. The
// mutating form is never invoked. The fixture is the only working directory
// and the only --workspace.
func TestCascadeCloseClassifyOnlySmoke(t *testing.T) {
	cascadeSmokeRequireTools(t)

	repo := cascadeSmokeRepoRoot(t)
	evidenceBefore, err := cascadeSmokeEvidenceSnapshot(repo)
	if err != nil {
		t.Fatalf("snapshot real docs/closure/evidence: %v", err)
	}
	t.Cleanup(func() {
		evidenceAfter, snapErr := cascadeSmokeEvidenceSnapshot(repo)
		if snapErr != nil {
			t.Errorf("snapshot real docs/closure/evidence after test: %v", snapErr)
			return
		}
		if !reflect.DeepEqual(evidenceBefore, evidenceAfter) {
			t.Errorf("real repository docs/closure/evidence changed during smoke test: before=%v after=%v", evidenceBefore, evidenceAfter)
		}
	})

	fixture := cascadeSmokeBuildFixture(t)
	scratch := t.TempDir()

	first := cascadeSmokeRun("autoharness", fixture, scratch, cascadeSmokeClassifyArgs(fixture)...)
	cascadeSmokeMustComplete(t, "first --classify-only", first)
	rep := cascadeSmokeParseReport(t, first)
	if rep.Mode != "classify_only" {
		t.Fatalf("mode = %q, want classify_only", rep.Mode)
	}
	if !cascadeSmokeVerdicts[rep.ClassifierVerdict] {
		t.Fatalf("classifier_verdict %q is not a documented value\nstdout:\n%s", rep.ClassifierVerdict, first.stdout)
	}
	// The fixture is built to select CASCADE, which the contract maps to exit 0.
	if rep.ClassifierVerdict != "CASCADE" {
		t.Fatalf("classifier_verdict = %q, want CASCADE for this fixture\nstdout:\n%s", rep.ClassifierVerdict, first.stdout)
	}
	if first.code != 0 {
		t.Fatalf("exit code = %d, want 0 (CASCADE selected)\nstdout:\n%s\nstderr:\n%s", first.code, first.stdout, first.stderr)
	}
	if !cascadeSmokeDocumentedCodes[first.code] {
		t.Fatalf("exit code %d is not a documented classify-only code", first.code)
	}
	if rep.ExitCode != first.code {
		t.Fatalf("json exit_code = %d, process exit = %d", rep.ExitCode, first.code)
	}
	if rep.MutationPossible != "no" {
		t.Fatalf("mutation_possible = %q, want no", rep.MutationPossible)
	}
	if len(rep.Failures) != 0 {
		t.Fatalf("failures = %v, want none", rep.Failures)
	}

	if rep.EvidencePath == "" {
		t.Fatalf("evidence_path is empty\nstdout:\n%s", first.stdout)
	}
	evidencePath := filepath.FromSlash(rep.EvidencePath)
	if !filepath.IsAbs(evidencePath) {
		t.Fatalf("evidence_path %q is not absolute", rep.EvidencePath)
	}
	evidence := cascadeSmokeResolve(t, evidencePath)
	rel, relErr := filepath.Rel(cascadeSmokeResolve(t, fixture), evidence)
	want := filepath.Join("docs", "closure", "evidence", cascadeSmokeEvidence)
	if relErr != nil || rel != want {
		t.Fatalf("evidence record %s is not at %s under fixture %s (rel=%q, err=%v)", evidence, want, fixture, rel, relErr)
	}

	record, err := os.ReadFile(evidence)
	if err != nil {
		t.Fatalf("read evidence record: %v", err)
	}
	var rec cascadeSmokeEvidenceRecord
	if err := json.Unmarshal(record, &rec); err != nil {
		t.Fatalf("parse evidence record: %v", err)
	}
	if rec.Phase != "pre_close" || rec.ShipmentID != cascadeSmokeShipment || rec.FeatureID != cascadeSmokeFeature {
		t.Fatalf("evidence record = phase %q shipment %q feature %q, want pre_close %s %s", rec.Phase, rec.ShipmentID, rec.FeatureID, cascadeSmokeShipment, cascadeSmokeFeature)
	}

	second := cascadeSmokeRun("autoharness", fixture, scratch, cascadeSmokeClassifyArgs(fixture)...)
	cascadeSmokeMustComplete(t, "second --classify-only", second)
	if second.code != 2 {
		t.Fatalf("second plain --classify-only exit = %d, want 2 (no-clobber)\nstdout:\n%s\nstderr:\n%s", second.code, second.stdout, second.stderr)
	}
	if rep2 := cascadeSmokeParseReport(t, second); rep2.ExitCode != 2 {
		t.Fatalf("second call json exit_code = %d, want 2", rep2.ExitCode)
	}
	after, err := os.ReadFile(evidence)
	if err != nil {
		t.Fatalf("re-read evidence record: %v", err)
	}
	if !bytes.Equal(record, after) {
		t.Fatalf("second --classify-only modified the first evidence record")
	}
}

// cascadeSmokeRequireTools skips, never passes vacuously, when the real
// autoharness or backlogit binaries are unavailable.
func cascadeSmokeRequireTools(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"autoharness", "backlogit"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("skipping cascade-close smoke: %s is not on PATH: %v", bin, err)
		}
	}
	scratch := t.TempDir()
	help := cascadeSmokeRun("autoharness", scratch, scratch, "shipment", "--help")
	if help.timedOut || help.runErr != nil || help.code != 0 {
		t.Skipf("skipping cascade-close smoke: `autoharness shipment --help` failed (timedOut=%t, code=%d, err=%v)\n%s%s", help.timedOut, help.code, help.runErr, help.stdout, help.stderr)
	}
}

// cascadeSmokeBuildFixture creates the smallest backlogit workspace that
// autoharness classifies as CASCADE. The covering feature is a manifest member
// along with its task, which makes the feature a fully covered root.
func cascadeSmokeBuildFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	scratch := t.TempDir()
	steps := []struct {
		args []string
		want string
	}{
		{[]string{"init"}, ""},
		{[]string{"add", "--type", "feature", "--title", "Smoke feature", "--status", "active"}, cascadeSmokeFeature},
		{[]string{"add", "--type", "task", "--title", "Smoke task", "--parent", cascadeSmokeFeature, "--status", "active"}, cascadeSmokeTask},
		{[]string{"shipment", "create", "--title", "Smoke shipment", "--items", cascadeSmokeFeature + "," + cascadeSmokeTask}, cascadeSmokeShipment},
	}
	for _, step := range steps {
		out := cascadeSmokeBacklogit(t, dir, scratch, step.args...)
		if step.want != "" && !strings.Contains(out, step.want) {
			t.Fatalf("backlogit %v: output does not contain %q:\n%s", step.args, step.want, out)
		}
	}
	// autoharness fails closed unless .backlog/archive exists, and backlogit
	// init does not create it.
	if err := os.MkdirAll(filepath.Join(dir, ".backlog", "archive"), 0o755); err != nil {
		t.Fatalf("create archive dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".autoharness"), 0o755); err != nil {
		t.Fatalf("create .autoharness dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".autoharness", "backlog-registry.yaml"), []byte(cascadeSmokeRegistry), 0o644); err != nil {
		t.Fatalf("write backlog registry: %v", err)
	}
	return dir
}

// cascadeSmokeBacklogit runs backlogit against dir and returns its stdout.
func cascadeSmokeBacklogit(t *testing.T, dir, scratch string, args ...string) string {
	t.Helper()
	full := append(append([]string{}, args...), "--cwd", dir, "--log-level", "error", "--no-update-check")
	o := cascadeSmokeRun("backlogit", dir, scratch, full...)
	cascadeSmokeMustComplete(t, "backlogit "+strings.Join(args, " "), o)
	if o.code != 0 {
		t.Fatalf("backlogit %v: exit %d\nstdout:\n%s\nstderr:\n%s", args, o.code, o.stdout, o.stderr)
	}
	return string(o.stdout)
}

// cascadeSmokeClassifyArgs returns the plain --classify-only argument list. It
// never includes --replace-pre-close or the mutating form.
func cascadeSmokeClassifyArgs(workspace string) []string {
	return []string{
		"shipment", "cascade-close", "--classify-only",
		"--shipment", cascadeSmokeShipment,
		"--feature", cascadeSmokeFeature,
		"--sha", cascadeSmokeSHA,
		"--workspace", workspace,
		"--json",
	}
}

// cascadeSmokeRun runs name with args in dir. TEMP, TMP, and TMPDIR point at
// tmp so child scratch output stays inside the test's temp directory.
func cascadeSmokeRun(name, dir, tmp string, args ...string) cascadeSmokeOutcome {
	ctx, cancel := context.WithTimeout(context.Background(), cascadeSmokeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = cascadeSmokeEnv(tmp)
	// Without WaitDelay, a grandchild holding the output pipes can block Wait
	// and keep the t.TempDir() directory locked after a timeout kill.
	cmd.WaitDelay = cascadeSmokeWaitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := cascadeSmokeOutcome{stdout: stdout.Bytes(), stderr: stderr.Bytes()}
	var exitErr *exec.ExitError
	switch {
	case ctx.Err() != nil:
		out.timedOut = true
	case errors.As(err, &exitErr):
		out.code = exitErr.ExitCode()
	case err != nil:
		out.runErr = err
	}
	return out
}

// cascadeSmokeMustComplete fails the test when a run timed out or could not
// start. Exit codes are left to the caller.
func cascadeSmokeMustComplete(t *testing.T, what string, o cascadeSmokeOutcome) {
	t.Helper()
	switch {
	case o.timedOut:
		t.Fatalf("%s: timed out after %s (not an exit code)\nstdout:\n%s\nstderr:\n%s", what, cascadeSmokeTimeout, o.stdout, o.stderr)
	case o.runErr != nil:
		t.Fatalf("%s: could not run: %v", what, o.runErr)
	}
}

// cascadeSmokeEnv inherits the current environment with the temp-directory
// variables redirected to tmp.
func cascadeSmokeEnv(tmp string) []string {
	env := make([]string, 0, len(os.Environ())+3)
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		switch strings.ToUpper(key) {
		case "TEMP", "TMP", "TMPDIR":
			continue
		}
		env = append(env, kv)
	}
	return append(env, "TEMP="+tmp, "TMP="+tmp, "TMPDIR="+tmp)
}

// cascadeSmokeParseReport decodes the --json object from a run's stdout.
func cascadeSmokeParseReport(t *testing.T, o cascadeSmokeOutcome) cascadeSmokeReport {
	t.Helper()
	var rep cascadeSmokeReport
	if err := json.Unmarshal(o.stdout, &rep); err != nil {
		t.Fatalf("stdout is not the --json object: %v\nstdout:\n%s\nstderr:\n%s", err, o.stdout, o.stderr)
	}
	return rep
}

// cascadeSmokeResolve returns the symlink-resolved form of path. Both sides of
// each path comparison go through it because t.TempDir paths can use 8.3 short
// names while the tool reports long names.
func cascadeSmokeResolve(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("resolve %s: %v", path, err)
	}
	return resolved
}

// cascadeSmokeRepoRoot walks up from the package directory to the module root.
func cascadeSmokeRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// cascadeSmokeEvidenceSnapshot maps each file under the real repository's
// docs/closure/evidence to its SHA-256 digest. A missing directory yields an
// empty snapshot.
func cascadeSmokeEvidenceSnapshot(root string) (map[string]string, error) {
	dir := filepath.Join(root, "docs", "closure", "evidence")
	snap := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		sum := sha256.Sum256(data)
		snap[filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return snap, nil
}
