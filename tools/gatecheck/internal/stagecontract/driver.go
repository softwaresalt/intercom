// Package stagecontract holds the fixture corpus, driver and detector for the
// Stage default-branch contract (properties P1-P6).
package stagecontract

import "errors"

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
	panic("not implemented: CaseResult.Failure report")
}

// StubDetector is the placeholder detector that always returns ErrNotImplemented.
type StubDetector struct{}

// Detect always fails with ErrNotImplemented.
func (StubDetector) Detect(Fixture) (Verdict, string, error) {
	panic("not implemented: StubDetector.Detect")
}

// RunSuite loads the manifest at manifestPath and runs one case per fixture
// through d. A malformed, truncated, unreadable or empty manifest is an error.
func RunSuite(manifestPath string, d Detector) ([]CaseResult, error) {
	panic("not implemented: RunSuite")
}
