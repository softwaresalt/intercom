//go:build windows

package writepath

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unicode/utf16"
)

// fsctlSetReparsePoint is FSCTL_SET_REPARSE_POINT and ioReparseTagMountPoint is
// IO_REPARSE_TAG_MOUNT_POINT, the tag of an NTFS junction.
const (
	fsctlSetReparsePoint   = 0x000900A4
	ioReparseTagMountPoint = 0xA0000003
)

// nativeJunction makes link a junction to target with a direct reparse-point
// call, so no shell ever re-parses a path that holds metacharacters. link must
// not exist yet.
func nativeJunction(target, link string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	if err := os.Mkdir(link, 0o755); err != nil {
		return err
	}
	if err := setMountPoint(link, abs); err != nil {
		_ = os.Remove(link)
		return err
	}
	return nil
}

// junctionNames returns the NT substitute name and the user-visible print name
// for a junction target. \\?\ extended-length and UNC spellings are normalized
// first, because the substitute name must be \??\C:\... or \??\UNC\server\share
// and never carry a literal \??\\\?\ or \??\\\server prefix.
func junctionNames(target string) (substitute, printName string) {
	switch {
	case hasPrefixFold(target, `\\?\UNC\`):
		target = `\\` + target[len(`\\?\UNC\`):]
	case hasPrefixFold(target, `\\?\`):
		target = target[len(`\\?\`):]
	}
	if strings.HasPrefix(target, `\\`) {
		return `\??\UNC\` + target[2:], target
	}
	return `\??\` + target, target
}

// hasPrefixFold reports whether s starts with prefix, ignoring ASCII case.
func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

func setMountPoint(link, target string) error {
	subName, printNameStr := junctionNames(target)
	sub := utf16.Encode([]rune(subName))
	printName := utf16.Encode([]rune(printNameStr))
	pathBytes := (len(sub) + 1 + len(printName) + 1) * 2
	buf := make([]byte, 8+8+pathBytes)
	binary.LittleEndian.PutUint32(buf[0:], ioReparseTagMountPoint)
	binary.LittleEndian.PutUint16(buf[4:], uint16(8+pathBytes))
	binary.LittleEndian.PutUint16(buf[8:], 0)
	binary.LittleEndian.PutUint16(buf[10:], uint16(len(sub)*2))
	binary.LittleEndian.PutUint16(buf[12:], uint16((len(sub)+1)*2))
	binary.LittleEndian.PutUint16(buf[14:], uint16(len(printName)*2))
	off := 16
	for _, u := range sub {
		binary.LittleEndian.PutUint16(buf[off:], u)
		off += 2
	}
	off += 2
	for _, u := range printName {
		binary.LittleEndian.PutUint16(buf[off:], u)
		off += 2
	}

	name, err := syscall.UTF16PtrFromString(link)
	if err != nil {
		return err
	}
	h, err := syscall.CreateFile(name, syscall.GENERIC_WRITE, 0, nil, syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return fmt.Errorf("open %s: %w", link, err)
	}
	defer func() { _ = syscall.CloseHandle(h) }()
	var returned uint32
	if err := syscall.DeviceIoControl(h, fsctlSetReparsePoint, &buf[0], uint32(len(buf)), nil, 0, &returned, nil); err != nil {
		return fmt.Errorf("set reparse point on %s: %w", link, err)
	}
	return nil
}

func TestJunctionNames(t *testing.T) {
	cases := []struct {
		name, target, wantSub, wantPrint string
	}{
		{"plain", `C:\work\dir`, `\??\C:\work\dir`, `C:\work\dir`},
		{"extended drive", `\\?\C:\work\dir`, `\??\C:\work\dir`, `C:\work\dir`},
		{"extended UNC", `\\?\UNC\srv\share\dir`, `\??\UNC\srv\share\dir`, `\\srv\share\dir`},
		{"extended UNC lowercase", `\\?\unc\srv\share\dir`, `\??\UNC\srv\share\dir`, `\\srv\share\dir`},
		{"UNC", `\\srv\share\dir`, `\??\UNC\srv\share\dir`, `\\srv\share\dir`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sub, printName := junctionNames(tc.target)
			if sub != tc.wantSub || printName != tc.wantPrint {
				t.Errorf("junctionNames(%q) = (%q, %q), want (%q, %q)", tc.target, sub, printName, tc.wantSub, tc.wantPrint)
			}
		})
	}
}

// TestNativeJunctionExtendedTarget builds a real junction from a \\?\ target and
// checks it resolves like the plain form.
func TestNativeJunctionExtendedTarget(t *testing.T) {
	target := t.TempDir()
	extended := `\\?\` + target
	link := filepath.Join(t.TempDir(), "ext-link")
	if err := nativeJunction(extended, link); err != nil {
		t.Fatalf("nativeJunction(%q): %v", extended, err)
	}
	if err := os.WriteFile(filepath.Join(target, "marker"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(link, "marker")); err != nil {
		t.Errorf("junction made from an extended target does not resolve: %v", err)
	}
}

// TestNativeJunctionUNCTarget is SKIPPED, not passed, when no UNC target can be
// constructed: the administrative share of the local drive is the only one
// available without external setup.
func TestNativeJunctionUNCTarget(t *testing.T) {
	dir := t.TempDir()
	vol := filepath.VolumeName(dir)
	if len(vol) != 2 || vol[1] != ':' {
		t.Skipf("temp dir %q has no drive letter; cannot build a UNC path", dir)
	}
	hostname, err := os.Hostname()
	if err != nil {
		t.Skipf("hostname unavailable: %v", err)
	}
	unc := `\\` + hostname + `\` + vol[:1] + `$` + dir[len(vol):]
	if _, err := os.Stat(unc); err != nil {
		t.Skipf("administrative share %q not reachable from this host: %v", unc, err)
	}
	link := filepath.Join(t.TempDir(), "unc-link")
	if err := nativeJunction(unc, link); err != nil {
		t.Fatalf("nativeJunction(%q): %v", unc, err)
	}
	if _, err := os.Stat(link); err != nil {
		t.Skipf("Windows does not resolve a junction with a UNC target (%v); the substitute-name form is asserted by TestJunctionNames", err)
	}
}
