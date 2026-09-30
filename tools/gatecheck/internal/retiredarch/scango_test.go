package retiredarch

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func goFixtureDir(root, sub string) string {
	return filepath.Join(root, "scripts", "testdata", sub)
}

func readDirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	return names, nil
}

func TestScanGo_MatchesGolden(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)

	for _, row := range g.GoFixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			dir := filepath.Join(root, "scripts", "testdata", filepath.Dir(row.Name))
			path := filepath.Join(dir, filepath.Base(row.Name))

			masked, err := scanGo(path, true)
			if err != nil {
				t.Fatalf("scanGo(mask=true, %s): %v", path, err)
			}
			wantMasked := derootifyAll(row.MaskedFindings, g.RootPlaceholder, filepath.ToSlash(root))
			maskedOK := reflect.DeepEqual(masked, wantMasked) || (len(masked) == 0 && len(wantMasked) == 0)
			if !maskedOK {
				t.Fatalf("scanGo(mask=true) = %v, want %v", masked, wantMasked)
			}

			unmasked, err := scanGo(path, false)
			if err != nil {
				t.Fatalf("scanGo(mask=false, %s): %v", path, err)
			}
			wantUnmasked := derootifyAll(row.UnmaskedFindings, g.RootPlaceholder, filepath.ToSlash(root))
			unmaskedOK := reflect.DeepEqual(unmasked, wantUnmasked) || (len(unmasked) == 0 && len(wantUnmasked) == 0)
			if !unmaskedOK {
				t.Fatalf("scanGo(mask=false) = %v, want %v", unmasked, wantUnmasked)
			}
		})
	}
}

// TestScanGo_DifferentialSuite specifically walks the
// retiredgo-differential fixture suite, which is reject-only and always
// scanned unmasked (mask=false) via the "go-unmasked" engine override.
func TestScanGo_DifferentialSuite(t *testing.T) {
	root := repoRoot(t)
	dir := goFixtureDir(root, "retiredgo-differential")
	entries, err := readDirNames(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	sort.Strings(entries)
	if len(entries) == 0 {
		t.Fatalf("no fixtures discovered in %s", dir)
	}
	for _, name := range entries {
		name := name
		t.Run(name, func(t *testing.T) {
			findings, err := scanGo(filepath.Join(dir, name), false)
			if err != nil {
				t.Fatalf("scanGo: %v", err)
			}
			if len(findings) == 0 {
				t.Fatalf("expected the go-differential fixture %s to be rejected (unmasked scan), got clean", name)
			}
		})
	}
}
