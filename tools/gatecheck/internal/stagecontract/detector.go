package stagecontract

import "errors"

// ErrNoVerdict is returned when a document has no structure the requested
// property can be evaluated against; unknown content yields no verdict.
var ErrNoVerdict = errors.New("stage contract: no verdict")

// Construct is one member of the closed construct set, mapped to a property.
type Construct struct {
	Property    string
	Description string
}

// ContractDetector evaluates the closed set of Stage contract properties P1-P6.
type ContractDetector struct{}

// NewContractDetector returns the real Stage contract detector.
func NewContractDetector() *ContractDetector {
	panic("not implemented: closed-set P1-P6 detector")
}

// Constructs returns the enumerated closed construct set.
func (*ContractDetector) Constructs() []Construct {
	panic("not implemented: closed-set P1-P6 detector")
}

// Detect evaluates f.Property over the document structure of f.Content.
func (*ContractDetector) Detect(Fixture) (Verdict, string, error) {
	panic("not implemented: closed-set P1-P6 detector")
}
