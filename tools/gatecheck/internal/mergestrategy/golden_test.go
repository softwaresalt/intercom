package mergestrategy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// mergestrategyGolden mirrors testdata/mergestrategy_golden.json, captured
// from the Python evaluate_json body at the M3-T2 parent commit (see
// docs/plans/evidence/2026-09-28-gate-engine-go-migration/m3.md).
type mergestrategyGolden struct {
	PythonVersion string       `json:"python_version"`
	FixtureCases  []goldenCase `json:"fixture_cases"`
	InlineCases   []goldenCase `json:"inline_cases"`
}

type goldenCase struct {
	Name          string `json:"name"`
	Content       string `json:"content"`
	ContentBase64 string `json:"content_base64"`
	Verdict       string `json:"verdict"`
	Reason        string `json:"reason"`
	ExitCode      int    `json:"exit_code"`
	IsInvalidUTF8 bool   `json:"is_invalid_utf8"`
}

func loadGolden(t *testing.T) mergestrategyGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "mergestrategy_golden.json"))
	if err != nil {
		t.Fatalf("read mergestrategy_golden.json: %v", err)
	}
	var g mergestrategyGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal mergestrategy_golden.json: %v", err)
	}
	return g
}

// caseBytes returns the raw payload bytes for a golden case, decoding
// content_base64 when present (the invalid-UTF-8 case) or the content
// string otherwise.
func caseBytes(t *testing.T, c goldenCase) []byte {
	t.Helper()
	if c.ContentBase64 != "" {
		data, err := base64.StdEncoding.DecodeString(c.ContentBase64)
		if err != nil {
			t.Fatalf("decode content_base64 for %s: %v", c.Name, err)
		}
		return data
	}
	return []byte(c.Content)
}

// runGoldenCase evaluates c's payload via a temp file (the "by path"
// transport) and asserts the golden's verdict/exit_code -- and, unless
// the case is exempted (ED-8: the NaN-literal reason text, and the two
// trailing-data cases' Python-specific JSONDecodeError message text),
// also the golden's exact reason text.
func runGoldenCase(t *testing.T, c goldenCase) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.json")
	if err := os.WriteFile(path, caseBytes(t, c), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	if c.IsInvalidUTF8 {
		verdict, _, code, readErr := Evaluate(path, nil)
		if verdict != "" {
			t.Errorf("%s: expected empty verdict (raw read-error path), got %q", c.Name, verdict)
		}
		if code != c.ExitCode {
			t.Errorf("%s: exit code = %d, want %d", c.Name, code, c.ExitCode)
		}
		if readErr == "" {
			t.Errorf("%s: expected a non-empty readErr (ED-2 synthetic ::error:: line)", c.Name)
		}
		return
	}

	verdict, reason, code, readErr := Evaluate(path, nil)
	if readErr != "" {
		t.Fatalf("%s: unexpected readErr: %s", c.Name, readErr)
	}
	if verdict != c.Verdict {
		t.Errorf("%s: verdict = %s, want %s", c.Name, verdict, c.Verdict)
	}
	if code != c.ExitCode {
		t.Errorf("%s: exit code = %d, want %d", c.Name, code, c.ExitCode)
	}

	switch c.Name {
	case "nan-literal", "trailing-data-after-object", "two-concatenated-objects":
		// ED-8: Go's encoding/json rejects the NaN literal outright (a
		// decode error, unlike Python's permissive json.loads), and both
		// Go's and Python's "trailing data" messages are
		// implementation-specific free text. Only the verdict and exit
		// code are characterized as parity for these three cases; the
		// reason text is deliberately exempted.
		if reason == "" {
			t.Errorf("%s: expected a non-empty reason", c.Name)
		}
	default:
		if reason != c.Reason {
			t.Errorf("%s: reason = %q, want %q", c.Name, reason, c.Reason)
		}
	}
}

func TestEvaluate_FixtureCases_ByPath(t *testing.T) {
	g := loadGolden(t)
	for _, c := range g.FixtureCases {
		t.Run(c.Name, func(t *testing.T) {
			runGoldenCase(t, c)
		})
	}
}

func TestEvaluate_InlineCases_ByPath(t *testing.T) {
	g := loadGolden(t)
	for _, c := range g.InlineCases {
		t.Run(c.Name, func(t *testing.T) {
			runGoldenCase(t, c)
		})
	}
}

// TestEvaluate_StdinTransport re-runs every golden case through the "-"
// stdin transport (mirroring evaluate_response's own in-memory-response
// path in the bash wrapper), asserting the same verdict/exit_code a
// by-path evaluation produced, so a transport-specific regression cannot
// hide behind the file-path tests above.
func TestEvaluate_StdinTransport(t *testing.T) {
	g := loadGolden(t)
	all := append(append([]goldenCase{}, g.FixtureCases...), g.InlineCases...)
	for _, c := range all {
		t.Run(c.Name, func(t *testing.T) {
			data := caseBytes(t, c)
			if c.IsInvalidUTF8 {
				verdict, _, code, readErr := Evaluate("-", bytes.NewReader(data))
				if verdict != "" || readErr == "" {
					t.Errorf("%s (stdin): expected raw read-error path", c.Name)
				}
				if code != c.ExitCode {
					t.Errorf("%s (stdin): exit code = %d, want %d", c.Name, code, c.ExitCode)
				}
				return
			}
			verdict, _, code, readErr := Evaluate("-", bytes.NewReader(data))
			if readErr != "" {
				t.Fatalf("%s (stdin): unexpected readErr: %s", c.Name, readErr)
			}
			if verdict != c.Verdict {
				t.Errorf("%s (stdin): verdict = %s, want %s", c.Name, verdict, c.Verdict)
			}
			if code != c.ExitCode {
				t.Errorf("%s (stdin): exit code = %d, want %d", c.Name, code, c.ExitCode)
			}
		})
	}
}
