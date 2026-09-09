// Package retiredgo lives under scripts/testdata, so the Go toolchain ignores it.
package retiredgo

type ConfigWithTag struct {
	Channel string `toml:"channel_id"`
}
