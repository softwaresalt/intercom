package stagecontract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestWellFormed(t *testing.T) {
	m, err := loadManifest(filepath.Join("testdata", "manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(m.Fixtures) == 0 {
		t.Fatal("manifest has no fixtures")
	}

	properties := []string{"P1", "P2", "P3", "P4", "P5", "P6"}
	verdicts := map[string]map[string]bool{}
	for _, p := range properties {
		verdicts[p] = map[string]bool{}
	}
	referenced := map[string]bool{}
	ids := map[string]bool{}

	for _, e := range m.Fixtures {
		if e.ID == "" || ids[e.ID] {
			t.Errorf("fixture id %q is empty or duplicated", e.ID)
		}
		ids[e.ID] = true
		if e.Reason == "" {
			t.Errorf("fixture %s: empty reason", e.ID)
		}
		if _, ok := verdicts[e.Property]; !ok {
			t.Errorf("fixture %s: unknown property %q", e.ID, e.Property)
			continue
		}
		if e.Verdict != "violating" && e.Verdict != "conforming" {
			t.Errorf("fixture %s: unknown verdict %q", e.ID, e.Verdict)
			continue
		}
		verdicts[e.Property][e.Verdict] = true
		referenced[e.File] = true
		if _, err := os.Stat(filepath.Join("testdata", e.File)); err != nil {
			t.Errorf("fixture %s: referenced file missing: %v", e.ID, err)
		}
	}

	for _, p := range properties {
		for _, v := range []string{"violating", "conforming"} {
			if !verdicts[p][v] {
				t.Errorf("property %s has no %s fixture", p, v)
			}
		}
	}

	files, err := filepath.Glob(filepath.Join("testdata", "*.md"))
	if err != nil {
		t.Fatalf("glob fixtures: %v", err)
	}
	for _, f := range files {
		if !referenced[filepath.Base(f)] {
			t.Errorf("unmapped fixture file %s", filepath.Base(f))
		}
	}
}
