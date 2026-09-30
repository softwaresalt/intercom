// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 8387758F (029-S decision delta D-1, folded into shipment 044-F / M1-T1):
// the masker leaves a raw-string literal's content fully VISIBLE when that
// content, taken as text, matches Go's struct-tag grammar end to end — real
// struct tags carry config keys the gate must not hide. This fixture proves
// that same rule also lets a qualified write selector hide in plain sight
// when it is the VALUE half of a tag-shaped key:"value" pair: the struct
// tag below is real (attached to a real exported field), so its content is
// unmasked, and the selector os.Remove must be reported as a finding at the
// tag's own source line.
package writepath

type StructTagSelector struct {
	Field int `x:"os.Remove"`
}
