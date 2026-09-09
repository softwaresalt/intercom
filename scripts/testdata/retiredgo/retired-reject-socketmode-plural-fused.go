// Package retiredgo lives under scripts/testdata, so the Go toolchain ignores it.
package retiredgo

// Post-015.005-T review (M-2 regression): a fused, unseparated, PLURAL
// identifier -- neither a case boundary (like SocketModes) nor an
// underscore (like socket_modes) breaks it into multiple components, and
// it is not the bare singular fused form (socketmode) either. Proves
// segment_whole()'s plural-suffix retry closes this combined evasion.
var socketmodes []string
