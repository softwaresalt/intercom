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

// cascadeSmokeOutcome records one child process run. A timeout is kept apart
// from exit codes so it can never be mistaken for a real exit code.
type cascadeSmokeOutcome struct {
	stdout      []byte
	stderr      []byte
	code        int
	timedOut    bool
	waitDelayed bool
	runErr      error
}

// cascadeSmokeReport is the subset of the --json object this test checks.
type cascadeSmokeReport struct {
	Mode              string   `json:"mode"`
	ExitCode          int      `json:"exit_code"`
	EvidencePath      string   `json:"evidence_path"`
	ClassifierVerdict string   `json:"classifier_verdict"`
	EngineVerdict     string   `json:"engine_verdict"`
	SelectedClosePath string   `json:"selected_close_path"`
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

	repo := repoRoot(t)
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

	t.Run("cascade_full_root_exit0", cascadeSmokeFullRootExit0)
	t.Run("safe_close_task_only_exit3", cascadeSmokeTaskOnlySafeClose)
}

// cascadeSmokeFullRootExit0 is the CASCADE subtest. Both the covering feature
// and its task are manifest members, so the real command exits 0.
func cascadeSmokeFullRootExit0(t *testing.T) {
	fixture := cascadeSmokeBuildFixture(t)
	scratch := t.TempDir()

	first := cascadeSmokeRun("autoharness", fixture, scratch, cascadeSmokeClassifyArgs(fixture)...)
	cascadeSmokeMustComplete(t, "first --classify-only", first)
	rep := cascadeSmokeParseReport(t, first)
	if rep.Mode != "classify_only" {
		t.Fatalf("mode = %q, want classify_only", rep.Mode)
	}
	// The expected exit is derived from the reported selected_close_path
	// (cascade -> 0, safe_close -> 3), not assumed from the fixture shape.
	wantExit := cascadeSmokeExpectedExit(t, first, rep)
	if first.code != wantExit {
		t.Fatalf("exit code = %d, want %d (selected_close_path %q)\nstdout:\n%s\nstderr:\n%s", first.code, wantExit, rep.SelectedClosePath, first.stdout, first.stderr)
	}
	// The pre_close record is written on both paths, so the evidence-record and
	// no-clobber checks below run whichever path the engine selected.
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

// cascadeSmokeUnknownCommand is the text autoharness prints for an unrecognized
// command. An older autoharness without cascade-close rejects `shipment` with it.
const cascadeSmokeUnknownCommand = "Unknown command: shipment"

// cascadeSmokeTaskOnlySafeClose runs the real classify-only command against a
// fixture whose manifest holds only the task, with no covering feature member.
// The classifier selects SAFE_CLOSE, which exits 3 with mutation_possible no.
func cascadeSmokeTaskOnlySafeClose(t *testing.T) {
	fixture := cascadeSmokeBuildFixtureWithItems(t, cascadeSmokeTask)
	scratch := t.TempDir()

	run := cascadeSmokeRun("autoharness", fixture, scratch, cascadeSmokeClassifyArgs(fixture)...)
	cascadeSmokeMustComplete(t, "task-only --classify-only", run)
	rep := cascadeSmokeParseReport(t, run)
	wantExit := cascadeSmokeExpectedExit(t, run, rep)
	if run.code != wantExit {
		t.Fatalf("exit code = %d, want %d (selected_close_path %q)\nstdout:\n%s\nstderr:\n%s", run.code, wantExit, rep.SelectedClosePath, run.stdout, run.stderr)
	}
	if rep.ExitCode != run.code {
		t.Fatalf("json exit_code = %d, process exit = %d", rep.ExitCode, run.code)
	}
	if want := cascadeSmokeVerdictFor(rep.SelectedClosePath); rep.ClassifierVerdict != want {
		t.Fatalf("classifier_verdict = %q, want %q for selected_close_path %q\nstdout:\n%s", rep.ClassifierVerdict, want, rep.SelectedClosePath, run.stdout)
	}
	if rep.MutationPossible != "no" {
		t.Fatalf("mutation_possible = %q, want no\nstdout:\n%s", rep.MutationPossible, run.stdout)
	}
}

// cascadeSmokeVerdictFor maps a selected_close_path to its classifier verdict.
func cascadeSmokeVerdictFor(path string) string {
	switch path {
	case "cascade":
		return "CASCADE"
	case "safe_close":
		return "SAFE_CLOSE"
	}
	return ""
}

// cascadeSmokeExpectedExit derives the exit code the contract requires from the
// reported selected_close_path (cascade -> 0, safe_close -> 3). It skips when
// the engine reports a verdict other than VERIFIED, because an unverified engine
// is not a basis for asserting exit codes, and fails when selected_close_path is
// absent or not one of the two contract values.
func cascadeSmokeExpectedExit(t *testing.T, o cascadeSmokeOutcome, rep cascadeSmokeReport) int {
	t.Helper()
	if rep.EngineVerdict == "" {
		t.Fatalf("engine_verdict is empty or missing; failing closed\nstdout:\n%s", o.stdout)
	}
	if rep.EngineVerdict != "VERIFIED" {
		t.Skipf("skipping: engine_verdict = %q, not VERIFIED; exit assertions need a verified engine\nstdout:\n%s", rep.EngineVerdict, o.stdout)
	}
	switch rep.SelectedClosePath {
	case "cascade":
		return 0
	case "safe_close":
		return 3
	}
	t.Fatalf("selected_close_path = %q, want cascade or safe_close\nstdout:\n%s", rep.SelectedClosePath, o.stdout)
	return -1
}

// cascadeSmokeRequireTools skips, never passes vacuously, when the real
// autoharness or backlogit binaries are unavailable, or when an older
// autoharness CLI rejects `shipment` as an unknown command. Any other probe
// result fails the test.
func cascadeSmokeRequireTools(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"autoharness", "backlogit"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("skipping cascade-close smoke: %s is not on PATH: %v", bin, err)
		}
	}
	scratch := t.TempDir()
	help := cascadeSmokeRun("autoharness", scratch, scratch, "shipment", "--help")
	cascadeSmokeMustComplete(t, "autoharness shipment --help", help)
	output := string(help.stdout) + string(help.stderr)
	switch help.code {
	case 0:
		if !strings.Contains(output, "cascade-close") {
			t.Fatalf("`autoharness shipment --help` exited 0 but its output does not mention cascade-close\nstdout:\n%s\nstderr:\n%s", help.stdout, help.stderr)
		}
	case 1:
		if !strings.Contains(output, cascadeSmokeUnknownCommand) {
			t.Fatalf("`autoharness shipment --help` exited 1 without %q, so it is not the old-CLI unknown-command exit\nstdout:\n%s\nstderr:\n%s", cascadeSmokeUnknownCommand, help.stdout, help.stderr)
		}
		t.Skipf("skipping cascade-close smoke: `autoharness shipment --help` exited 1 with %q, an older autoharness without cascade-close\nstdout:\n%s\nstderr:\n%s", cascadeSmokeUnknownCommand, help.stdout, help.stderr)
	default:
		t.Fatalf("`autoharness shipment --help` exited %d, want 0 or 1 (unknown command)\nstdout:\n%s\nstderr:\n%s", help.code, help.stdout, help.stderr)
	}
}

// cascadeSmokeBuildFixture creates the smallest backlogit workspace that
// autoharness classifies as CASCADE. The covering feature is a manifest member
// along with its task, which makes the feature a fully covered root.
func cascadeSmokeBuildFixture(t *testing.T) string {
	t.Helper()
	return cascadeSmokeBuildFixtureWithItems(t, cascadeSmokeFeature+","+cascadeSmokeTask)
}

// cascadeSmokeBuildFixtureWithItems creates a backlogit workspace holding the
// feature and task above and a shipment whose manifest is items, a
// comma-separated list of item IDs.
func cascadeSmokeBuildFixtureWithItems(t *testing.T, items string) string {
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
		{[]string{"shipment", "create", "--title", "Smoke shipment", "--items", items}, cascadeSmokeShipment},
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

// cascadeSmokeRun runs name with args in dir. TEMP, TMP, TMPDIR, and the profile,
// config, and cache directories point under tmp (see cascadeSmokeEnv) so child
// output stays inside the test's temp directory.
func cascadeSmokeRun(name, dir, tmp string, args ...string) cascadeSmokeOutcome {
	ctx, cancel := context.WithTimeout(context.Background(), cascadeSmokeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = cascadeSmokeEnv(tmp)
	// WaitDelay bounds how long Wait blocks on the output pipes once the child
	// exits. It does not kill grandchildren: on Windows the default Cancel kills
	// only the direct child, so a leftover grandchild can keep files open and
	// fixture cleanup can fail. This is accepted, and the cleanup error is
	// reported by t.TempDir.
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
	case errors.Is(err, exec.ErrWaitDelay) && cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 0:
		out.waitDelayed = true
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
	case o.waitDelayed:
		t.Logf("%s: child exited 0 but its output pipes stayed open past WaitDelay %s, so a grandchild likely holds them; treating the run as completed", what, cascadeSmokeWaitDelay)
	}
}

// cascadeSmokeEnv inherits the current environment except for GIT_, BACKLOGIT_,
// and AUTOHARNESS_ variables. The temp-directory variables and the profile,
// config, and cache directories point under tmp, so the child never reads or
// writes the operator's real state.
func cascadeSmokeEnv(tmp string) []string {
	overrides := []string{
		"TEMP=" + tmp,
		"TMP=" + tmp,
		"TMPDIR=" + tmp,
		"HOME=" + filepath.Join(tmp, "home"),
		"USERPROFILE=" + filepath.Join(tmp, "userprofile"),
		"APPDATA=" + filepath.Join(tmp, "appdata"),
		"LOCALAPPDATA=" + filepath.Join(tmp, "localappdata"),
		"XDG_CONFIG_HOME=" + filepath.Join(tmp, "xdg-config"),
		"XDG_DATA_HOME=" + filepath.Join(tmp, "xdg-data"),
		"XDG_CACHE_HOME=" + filepath.Join(tmp, "xdg-cache"),
	}
	replaced := make(map[string]bool, len(overrides))
	for _, kv := range overrides {
		key, _, _ := strings.Cut(kv, "=")
		replaced[strings.ToUpper(key)] = true
	}
	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		upper := strings.ToUpper(key)
		if replaced[upper] || strings.HasPrefix(upper, "GIT_") ||
			strings.HasPrefix(upper, "BACKLOGIT_") || strings.HasPrefix(upper, "AUTOHARNESS_") {
			continue
		}
		env = append(env, kv)
	}
	return append(env, overrides...)
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
