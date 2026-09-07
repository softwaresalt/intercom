// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test.
//
// This fixture's only purpose is to prove the detector masks comments and
// string literals before scanning: it mentions os.WriteFile and
// os.RemoveAll below, but only inside a comment and a string literal, never
// as real code, and must therefore be scored ACCEPT (clean).
package writepath

// A caller must never call os.WriteFile directly here; see the risk
// register for why.
func DescribeNonWritePrimitive() string {
	return "this string literal mentions os.RemoveAll but calls nothing"
}
