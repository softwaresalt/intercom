// Package stagecontract holds the fixture corpus, driver and detector for the
// Stage default-branch contract (properties P1-P6).
package stagecontract

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Verdict is the outcome a detector reports for one fixture.
type Verdict string

// Verdicts recognised by the manifest and the driver.
const (
	Conforming Verdict = "conforming"
	Violating  Verdict = "violating"
)

// ErrNotImplemented is returned by the stub detector until the real detector lands.
var ErrNotImplemented = errors.New("stage contract detector not implemented")

// Fixture is one contract-state document handed to a detector.
type Fixture struct {
	ID       string
	Property string
	Content  []byte
}

// Detector evaluates a fixture and returns its verdict and reason.
type Detector interface {
	Detect(f Fixture) (Verdict, string, error)
}

// CaseResult records the outcome of running one manifest fixture.
type CaseResult struct {
	ID              string
	ExpectedVerdict Verdict
	ExpectedReason  string
	ObservedVerdict Verdict
	ObservedReason  string
	Err             error
}

// Failure returns a report naming the fixture id and expected and observed
// verdicts, or the empty string when the case passed.
func (r CaseResult) Failure() string {
	if r.Err != nil {
		return fmt.Sprintf("fixture %s: expected verdict %q, detector error: %v", r.ID, r.ExpectedVerdict, r.Err)
	}
	if r.ObservedVerdict != r.ExpectedVerdict {
		return fmt.Sprintf("fixture %s: expected verdict %q, observed verdict %q", r.ID, r.ExpectedVerdict, r.ObservedVerdict)
	}
	if r.ObservedReason != r.ExpectedReason {
		return fmt.Sprintf("fixture %s: verdict %q matched but expected reason %q, observed reason %q", r.ID, r.ExpectedVerdict, r.ExpectedReason, r.ObservedReason)
	}
	return ""
}

// StubDetector is the placeholder detector that always returns ErrNotImplemented.
type StubDetector struct{}

// Detect always fails with ErrNotImplemented.
func (StubDetector) Detect(Fixture) (Verdict, string, error) {
	return "", "", ErrNotImplemented
}

type manifestEntry struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Property string `json:"property"`
	Verdict  string `json:"verdict"`
	Reason   string `json:"reason"`
}

type manifest struct {
	Fixtures []manifestEntry `json:"fixtures"`
}

// loadManifest reads and parses the manifest at path.
func loadManifest(path string) (manifest, error) {
	var m manifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, fmt.Errorf("read manifest: %w", err)
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("parse manifest: %w", err)
	}
	return m, nil
}

// RunSuite loads the manifest at manifestPath and runs one case per fixture
// through d. A malformed, truncated, unreadable or empty manifest is an error.
func RunSuite(manifestPath string, d Detector) ([]CaseResult, error) {
	m, err := loadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if len(m.Fixtures) == 0 {
		return nil, errors.New("manifest has no fixtures")
	}
	dir := filepath.Dir(manifestPath)
	results := make([]CaseResult, 0, len(m.Fixtures))
	for _, e := range m.Fixtures {
		v := Verdict(e.Verdict)
		if e.ID == "" || (v != Conforming && v != Violating) {
			return nil, fmt.Errorf("manifest entry %q has invalid id or verdict %q", e.ID, e.Verdict)
		}
		content, err := os.ReadFile(filepath.Join(dir, e.File))
		if err != nil {
			return nil, fmt.Errorf("fixture %s: read fixture file: %w", e.ID, err)
		}
		r := CaseResult{ID: e.ID, ExpectedVerdict: v, ExpectedReason: e.Reason}
		r.ObservedVerdict, r.ObservedReason, r.Err = d.Detect(Fixture{ID: e.ID, Property: e.Property, Content: content})
		results = append(results, r)
	}
	return results, nil
}
