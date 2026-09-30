package retiredarch

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScanTomlFallback_MatchesGolden_Fixtures(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	rootPosix := filepath.ToSlash(root)

	for _, row := range g.TomlFixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			path := filepath.Join(root, "scripts", "testdata", row.Name)
			got, err := scanTomlFallback(path)
			if err != nil {
				t.Fatalf("scanTomlFallback(%s): %v", path, err)
			}
			want := derootifyAll(row.FallbackFindings, g.RootPlaceholder, rootPosix)
			if row.Name == "retired-malformed-unparseable.toml" {
				assertFailClosedReject(t, got, want)
				return
			}
			ok := reflect.DeepEqual(got, want) || (len(got) == 0 && len(want) == 0)
			if !ok {
				t.Fatalf("scanTomlFallback(%s) = %v, want %v", path, got, want)
			}
		})
	}
}

func TestScanTomlFallback_MatchesGolden_Inline(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	rootPosix := filepath.ToSlash(root)
	dir := filepath.Join(root, "tools", "gatecheck", "internal", "retiredarch", "testdata", "_inline_toml_materialized")

	for name, row := range g.InlineTomlGoldens {
		name, row := name, row
		t.Run(name, func(t *testing.T) {
			var path string
			if name == bomCaseName {
				path = filepath.Join(dir, "bom-case.toml")
			} else {
				path = filepath.Join(dir, name+".toml")
			}
			got, err := scanTomlFallback(path)
			if err != nil {
				t.Fatalf("scanTomlFallback(%s): %v", path, err)
			}
			want := derootifyAll(row.FallbackFindings, g.RootPlaceholder, rootPosix)
			ok := reflect.DeepEqual(got, want) || (len(got) == 0 && len(want) == 0)
			if !ok {
				t.Fatalf("scanTomlFallback(%s) = %v, want %v", name, got, want)
			}
		})
	}
}

// TestDualEngineAgreement_Fixtures asserts the primary/fallback engines'
// AGREE VERDICT (both clean vs both non-empty) matches the golden for
// every real fixture, with the same malformed-fixture verdict-only carve
// out as the per-engine tests (ED-4).
func TestDualEngineAgreement_Fixtures(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)

	for _, row := range g.TomlFixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			path := filepath.Join(root, "scripts", "testdata", row.Name)
			primary := scanTomlPrimary(path)
			fallback, err := scanTomlFallback(path)
			if err != nil {
				t.Fatalf("scanTomlFallback(%s): %v", path, err)
			}
			gotAgree := (len(primary) == 0) == (len(fallback) == 0)
			if gotAgree != row.Agree {
				t.Fatalf("dual-engine agreement for %s = %v, want %v (primary=%v fallback=%v)", row.Name, gotAgree, row.Agree, primary, fallback)
			}
		})
	}
}

// TestDualEngineAgreement_Inline covers every new T1 inline case,
// including the ED-6/BOM sanctioned disagreements: a disagreement on any
// OTHER new case must fail this test rather than being silently skipped.
func TestDualEngineAgreement_Inline(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	dir := filepath.Join(root, "tools", "gatecheck", "internal", "retiredarch", "testdata", "_inline_toml_materialized")

	for name, row := range g.InlineTomlGoldens {
		name, row := name, row
		t.Run(name, func(t *testing.T) {
			var path string
			if name == bomCaseName {
				path = filepath.Join(dir, "bom-case.toml")
			} else {
				path = filepath.Join(dir, name+".toml")
			}
			primary := scanTomlPrimary(path)
			fallback, err := scanTomlFallback(path)
			if err != nil {
				t.Fatalf("scanTomlFallback(%s): %v", path, err)
			}
			gotAgree := (len(primary) == 0) == (len(fallback) == 0)
			switch name {
			case ed6CaseName:
				// Sanctioned ED-6 delta: golden says Python disagrees
				// (tomllib rejects, fallback accepts); the Go port's
				// primary engine accepts too (TOML 1.1), so Go's
				// dual-engine verdict is expected to AGREE (both clean),
				// which correctly differs from Python's own agree=false.
				if !gotAgree {
					t.Fatalf("ED-6 case: expected the Go primary/fallback engines to agree (both accept), got disagreement (primary=%v fallback=%v)", primary, fallback)
				}
			case bomCaseName:
				// C-6 BOM parity: Go primary rejects (fail-closed, same
				// as Python's tomllib), Go fallback accepts cleanly (same
				// pre-existing asymmetry as Python's own fallback), so Go
				// must reproduce the SAME disagreement Python recorded.
				if gotAgree != row.Agree {
					t.Fatalf("BOM case: dual-engine agreement = %v, want %v (primary=%v fallback=%v)", gotAgree, row.Agree, primary, fallback)
				}
			default:
				if gotAgree != row.Agree {
					t.Fatalf("dual-engine agreement for %s = %v, want %v (primary=%v fallback=%v)", name, gotAgree, row.Agree, primary, fallback)
				}
			}
		})
	}
}

// TestDualEngineAgreement_ReopenedTableForbiddenToken is a golden-
// independent regression guard for F-1 (adversarial review
// docs/closure/2026-09-30-gate-engine-m2-retired-arch-adversarial-
// review.md), mirroring
// TestScanTomlPrimary_ReopenedTableForbiddenToken: both engines must
// agree the reopened table's newly-added forbidden field is rejected
// (before the fix, the primary engine silently reported clean here while
// the fallback correctly rejected, i.e. a disagreement this test would
// have caught).
func TestDualEngineAgreement_ReopenedTableForbiddenToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reopened.toml")
	text := "[a.b]\nx = 1\n\n[a]\nchannel_id = 2\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	primary := scanTomlPrimary(path)
	fallback, err := scanTomlFallback(path)
	if err != nil {
		t.Fatalf("scanTomlFallback: %v", err)
	}
	if len(primary) == 0 {
		t.Fatalf("primary engine reported clean for a reopened table with a forbidden field; want a rejection")
	}
	if len(fallback) == 0 {
		t.Fatalf("fallback engine reported clean for a reopened table with a forbidden field; want a rejection")
	}
}
