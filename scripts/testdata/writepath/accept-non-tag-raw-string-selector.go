// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite this scan input existing purely to be read as
// text by the detector, not imported by anything.
//
// 8387758F (029-S decision delta D-1, folded into shipment 044-F / M1-T1):
// the counterpart to reject-struct-tag-selector.go. This raw string holds
// the SAME "x:\"os.Remove\"" text, but trailing free-form prose after the
// key:"value" pair means the content does NOT match the tag grammar end to
// end (struct_tag_re requires the whole content to be one or more
// whitespace-separated key:"value" pairs and nothing else). The masker
// therefore fully masks this raw string's content, and the detector must
// report this fixture ACCEPT (clean).
package writepath

func DescribeNonTagSelectorText() string {
	q := `x:"os.Remove" is only prose here, not a struct tag`
	return q
}
