package pathsafe

import (
	"path/filepath"
	"testing"
)

// TestStripUNCPrefixReturnsWindowsAbsolutePaths verifies the stripped result
// remains Windows-absolute for the cases whose postcondition requires it.
func TestStripUNCPrefixReturnsWindowsAbsolutePaths(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		want       string
		wantVolume string
	}{
		{name: "unc share root", in: `\\?\UNC\server\share`, want: `\\server\share`, wantVolume: `\\server\share`},
		{name: "unc share subpath", in: `\\?\UNC\server\share\ws\file`, want: `\\server\share\ws\file`, wantVolume: `\\server\share`},
		{name: "drive path", in: `\\?\C:\path`, want: `C:\path`, wantVolume: `C:`},
		{name: "lowercase unc prefix", in: `\\?\unc\server\share`, want: `\\server\share`, wantVolume: `\\server\share`},
		{name: "volume guid", in: `\\?\Volume{GUID}\path`, want: `\\?\Volume{GUID}\path`, wantVolume: `\\?\Volume{GUID}`},
		{name: "globalroot", in: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`, want: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`, wantVolume: `\\?\GLOBALROOT`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripUNCPrefix(tc.in)
			if got != tc.want {
				t.Fatalf("stripUNCPrefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
			if !filepath.IsAbs(got) {
				t.Fatalf("filepath.IsAbs(stripUNCPrefix(%q)) = false, want true; got %q", tc.in, got)
			}
			if volume := filepath.VolumeName(got); volume != tc.wantVolume {
				t.Fatalf("filepath.VolumeName(stripUNCPrefix(%q)) = %q, want %q", tc.in, volume, tc.wantVolume)
			}
		})
	}
}
