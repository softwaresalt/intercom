// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It is not part of
// the module's build (excluded from internal/** by not being imported and
// never referenced from real source); it exists purely as scan input.
package writepath

func CleanHelper(a, b int) int {
	return a + b
}
