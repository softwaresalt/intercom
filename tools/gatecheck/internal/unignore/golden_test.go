package unignore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// unignoreGolden mirrors testdata/unignore_golden.json, captured from the
// live check-unignore-regression.sh + Python heredoc engine at the M3-T1
// parent commit (see docs/plans/evidence/2026-09-28-gate-engine-go-
// migration/m3.md).
type unignoreGolden struct {
	PythonVersion string          `json:"python_version"`
	Denylist      []string        `json:"denylist"`
	SelfTest      streamCapture   `json:"self_test"`
	Scenarios     []scenarioRow   `json:"scenarios"`
	UnknownFlag   unknownFlagCase `json:"unknown_flag"`
}

type streamCapture struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

type scenarioRow struct {
	Name    string `json:"name"`
	Verdict string `json:"verdict"`
}

type unknownFlagCase struct {
	Args     []string `json:"args"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode int      `json:"exit_code"`
}

func loadGolden(t *testing.T) unignoreGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "unignore_golden.json"))
	if err != nil {
		t.Fatalf("read unignore_golden.json: %v", err)
	}
	var g unignoreGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal unignore_golden.json: %v", err)
	}
	return g
}

// repoRoot returns the repository top-level directory from this test
// package's own path (tools/gatecheck/internal/unignore), mirroring
// retiredarch's golden_test.go helper of the same name, so tests can
// exercise the real repository state without depending on the working
// directory `go test` happens to use.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return root
}
