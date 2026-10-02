// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.008-T (D-T5a, AC-D5.1, AC-D5.6): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. The masker leaves
// a raw-string literal fully visible whenever its whole content matches
// Go's struct-tag grammar, whether or not it is attached to a struct
// field. The literal below is an ordinary expression, not a field tag, so
// a go/ast walk that only inspected Field.Tag would silently drop this
// finding (a parity hazard for Unit E's G-3). The text gate reports the
// selector at the literal's own source line.
package writepath

func tagShapedExpr() string {
	q := `x:"os.Remove"`
	return q
}
