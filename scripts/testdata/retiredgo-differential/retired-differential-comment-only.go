// Package retiredgo lives under scripts/testdata, so the Go toolchain ignores it.
package retiredgo

// slack appears in this comment only. Under the PRODUCTION masked scan
// (scan_go(path, mask=True)) this comment is stripped and the token is
// invisible, which is what the sibling retired-accept-comment-only.go
// fixture (ACCEPT) proves. This file lives in the go-differential suite
// instead, whose engine_override runs scan_go(path, mask=False) -- masking
// intentionally bypassed -- so this same comment-only token IS visible and
// MUST be caught (manifest expects "reject"). The differential suite's
// purpose is exactly this: prove the masked production path is not
// false-green by re-scanning the identical content unmasked and requiring
// it to fail.
type Clean string
