package main

import "testing"

// TestRootCmd_ConfigFlagDefaultsToConfigDotToml covers unit E3's
// acceptance criterion (ii): the --config flag's DefValue equals
// "config.toml", aligning with config.DefaultConfigPath (the oracle's
// --config default, main.rs:60) rather than the previous empty-string
// default.
func TestRootCmd_ConfigFlagDefaultsToConfigDotToml(t *testing.T) {
	cmd := newRootCmd()
	flag := cmd.Flags().Lookup("config")
	if flag == nil {
		t.Fatal("--config flag not registered")
	}
	if got, want := flag.DefValue, "config.toml"; got != want {
		t.Errorf("--config DefValue = %q, want %q", got, want)
	}
}
