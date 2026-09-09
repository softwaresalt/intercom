// Package retiredgo lives under scripts/testdata, so the Go toolchain ignores it.
package retiredgo

// Non-tag raw backtick literal (e.g. a SQL query) that happens to mention a
// retired token in prose. This is NOT a struct tag literal -- its content
// does not match struct_tag_re -- so it must stay fully masked and remain
// an ACCEPT, proving the struct-tag unmasking seam does not regress into a
// general raw-string regex over unlexed text.
const retiredTokenInQuery = `SELECT * FROM t WHERE channel_id = ?`
