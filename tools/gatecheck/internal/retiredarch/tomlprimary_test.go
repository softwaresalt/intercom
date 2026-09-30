package retiredarch

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// ed6CaseName is the M2-T1 ED-6 sanctioned delta: BurntSushi/toml v1.6.0
// accepts TOML 1.1 syntax (a raw newline inside an inline table) that
// Python's tomllib (TOML 1.0 only) rejects. This is the ONLY inline case
// whose primary-engine golden expectation must NOT be asserted (plan
// C-8's ED-6).
const ed6CaseName = "ed6-newline-in-inline-table"

// bomCaseName's golden reject text is a literal Python message
// (tomllib's own parser error), which this port never reproduces
// verbatim (ED-4: only the reject VERDICT must match, not the prose).
const bomCaseName = "leading-utf8-bom"

func TestScanTomlPrimary_MatchesGolden_Fixtures(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	rootPosix := filepath.ToSlash(root)

	for _, row := range g.TomlFixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			path := filepath.Join(root, "scripts", "testdata", row.Name)
			got := scanTomlPrimary(path)
			want := derootifyAll(row.TomllibFindings, g.RootPlaceholder, rootPosix)
			if row.Name == "retired-malformed-unparseable.toml" {
				assertFailClosedReject(t, got, want)
				return
			}
			if !reflect.DeepEqual(got, want) && (len(got) != 0 || len(want) != 0) {
				t.Fatalf("scanTomlPrimary(%s) = %v, want %v", path, got, want)
			}
		})
	}
}

// assertFailClosedReject checks ONLY the fail-closed verdict (non-empty
// findings, one per ED-4: parse-error prose is exempt from byte parity --
// only the reject/accept verdict and exit-code-relevant shape matter).
func assertFailClosedReject(t *testing.T, got, want []string) {
	t.Helper()
	if len(want) == 0 {
		t.Fatalf("golden expected a clean parse but this is a known-malformed fixture")
	}
	if len(got) == 0 {
		t.Fatalf("scanTomlPrimary returned no findings for a fixture that must fail closed")
	}
}

func TestScanTomlPrimary_MatchesGolden_Inline(t *testing.T) {
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
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("materialized fixture missing: %v", err)
			}

			got := scanTomlPrimary(path)
			want := derootifyAll(row.TomllibFindings, g.RootPlaceholder, rootPosix)

			switch name {
			case ed6CaseName:
				// Sanctioned ED-6 delta: Python's tomllib rejects this
				// TOML-1.1-only shape; BurntSushi (TOML 1.1) accepts it.
				// Assert the Go engine's ACCEPT verdict explicitly rather
				// than silently skipping this case.
				if len(got) != 0 {
					t.Fatalf("ED-6 case: scanTomlPrimary = %v, want a clean (accepted) parse", got)
				}
			case bomCaseName:
				// ED-4: only the reject verdict, not the literal prose.
				assertFailClosedReject(t, got, want)
			default:
				if !reflect.DeepEqual(got, want) && (len(got) != 0 || len(want) != 0) {
					t.Fatalf("scanTomlPrimary(%s) = %v, want %v", name, got, want)
				}
			}
		})
	}
}

// TestScanTomlPrimary_BOMRawBytes exercises the exact captured raw BOM
// bytes (not the materialized-file convenience path), confirming the
// fail-closed reject holds when read from bytes indistinguishable from
// what the M2-T1 generator captured.
func TestScanTomlPrimary_BOMRawBytes(t *testing.T) {
	g := loadGolden(t)
	row, ok := g.InlineTomlGoldens[bomCaseName]
	if !ok {
		t.Fatalf("golden missing %q", bomCaseName)
	}
	if len(row.RawBytes) == 0 {
		t.Fatalf("golden %q missing raw_bytes", bomCaseName)
	}
	buf := make([]byte, len(row.RawBytes))
	for i, b := range row.RawBytes {
		buf[i] = byte(b)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "bom-case.toml")
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got := scanTomlPrimary(path)
	if len(got) == 0 {
		t.Fatalf("scanTomlPrimary on raw BOM bytes must fail closed, got clean result")
	}
}

// TestScanTomlPrimary_ReopenedTableForbiddenToken is a golden-independent
// regression guard for F-1 (adversarial review
// docs/closure/2026-09-30-gate-engine-m2-retired-arch-adversarial-
// review.md): a plain (non-array) table reopened later in the document
// via a second "[header]" line must still have its newly-added fields
// walked and checked against matchesForbiddenParts. Before the fix,
// walkTable's `if !ok { return nil }` branch on peekChild's rejection of
// the reopened header's own self-entry treated that as end-of-table even
// though this table's `channel_id` field was still unvisited, silently
// returning a clean (false-negative) result. This test writes the exact
// minimal reproduction directly (independent of the golden JSON/materialized
// fixture file this same case is ALSO captured under, as
// "reopened-table-forbidden-token", for the golden-driven
// MatchesGolden_Inline/DualEngineAgreement_Inline table tests).
func TestScanTomlPrimary_ReopenedTableForbiddenToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reopened.toml")
	text := "[a.b]\nx = 1\n\n[a]\nchannel_id = 2\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := scanTomlPrimary(path)
	if len(got) != 1 {
		t.Fatalf("scanTomlPrimary(reopened table with forbidden token) = %v, want exactly 1 finding for the reopened field a.channel_id", got)
	}
	want := "retired token 'channel_id' in TOML key path 'a.channel_id' (segment 'channel_id') (via sequence model)"
	if !strings.Contains(got[0], want) {
		t.Fatalf("scanTomlPrimary finding = %q, want it to contain %q", got[0], want)
	}
}
