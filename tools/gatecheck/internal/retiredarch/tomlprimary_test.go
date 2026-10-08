package retiredarch

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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

// tomlDesyncRow is one U1 desync shape (plan U1 scenario 1): doc carries a
// forbidden token at a key under a reopened/ancestor/interleaved table and
// mirror is the identical shape with that key renamed to a token-free key.
type tomlDesyncRow struct {
	name     string
	doc      string
	mirror   string
	keyPaths []string // expected finding key paths, in fallback (sorted) order
	segments []string // segment per keyPaths entry; nil means 'channel_id' for every entry
}

// tomlDesyncRows are the shapes that desync walkTable's lockstep cursor at
// main@f50dba4 (probed on windows/amd64 before this test was written). The
// plain ancestor reopen "[a.b] ... [a]" and the array-of-tables reopen
// "[[t]] ... [t.sub] ... [[t]]" do NOT desync (peekSelf and
// walkArrayOfTables already handle them), so per AC-1 they are recorded in
// the desync-free characterization table instead.
var tomlDesyncRows = []tomlDesyncRow{
	{
		name:     "non-contiguous-siblings",
		doc:      "[a.x]\nk = 1\n[b]\nz = 1\n[a.y]\nchannel_id = 2\n",
		mirror:   "[a.x]\nk = 1\n[b]\nz = 1\n[a.y]\nplain_key = 2\n",
		keyPaths: []string{"a.y.channel_id"},
	},
	{
		name:     "non-contiguous-sibling-header-token",
		doc:      "[a.x]\nk = 1\n[b]\nz = 1\n[a.channel_id]\nk = 1\n",
		mirror:   "[a.x]\nk = 1\n[b]\nz = 1\n[a.plain_key]\nk = 1\n",
		keyPaths: []string{"a.channel_id", "a.channel_id.k"},
		segments: []string{"channel_id", "k"},
	},
	{
		name:     "depth-3-ancestor-token",
		doc:      "[a.b.c]\nk = 1\n[a]\nchannel_id = 2\n[a.b.c.d]\nm = 1\n",
		mirror:   "[a.b.c]\nk = 1\n[a]\nplain_key = 2\n[a.b.c.d]\nm = 1\n",
		keyPaths: []string{"a.channel_id"},
	},
	{
		name:     "depth-3-descendant-token",
		doc:      "[a.b.c]\nk = 1\n[a]\nq = 2\n[a.b.c.d]\nchannel_id = 1\n",
		mirror:   "[a.b.c]\nk = 1\n[a]\nq = 2\n[a.b.c.d]\nplain_key = 1\n",
		keyPaths: []string{"a.b.c.d.channel_id"},
	},
	{
		name:     "interleaved-ancestor-reopen",
		doc:      "[a.b]\nx = 1\n[c]\ny = 1\n[a]\nchannel_id = 2\n",
		mirror:   "[a.b]\nx = 1\n[c]\ny = 1\n[a]\nplain_key = 2\n",
		keyPaths: []string{"a.channel_id"},
	},
	{
		name:     "interleaved-ancestor-reopen-two-tokens",
		doc:      "[a.b]\nchannel_id = 1\n[c]\ny = 1\n[a]\nchannel_id = 2\n",
		mirror:   "[a.b]\nplain_key = 1\n[c]\ny = 1\n[a]\nplain_key = 2\n",
		keyPaths: []string{"a.b.channel_id", "a.channel_id"},
	},
	{
		name:     "interleaved-array-of-tables-subtable",
		doc:      "[[t]]\nk = 1\n[u]\ny = 1\n[t.sub]\nchannel_id = 2\n",
		mirror:   "[[t]]\nk = 1\n[u]\ny = 1\n[t.sub]\nplain_key = 2\n",
		keyPaths: []string{"t.sub.channel_id"},
	},
}

// TestScanTomlPrimary_DesyncShapes_UseFallback is U1 scenario 1 (AC-1):
// each shape must first desync walkTable (so the row exercises the
// fallback path), then scanTomlPrimary must return exactly the key-path
// finding(s) for the token document and zero findings for the mirror,
// never the fail-closed desync parse-error text.
func TestScanTomlPrimary_DesyncShapes_UseFallback(t *testing.T) {
	for _, row := range tomlDesyncRows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			for _, doc := range []string{row.doc, row.mirror} {
				var data map[string]interface{}
				meta, err := toml.Decode(doc, &data)
				if err != nil {
					t.Fatalf("toml.Decode(%q): %v", doc, err)
				}
				var partial []string
				werr := walkTable("p", nil, data, &tomlCursor{keys: meta.Keys()}, &partial)
				if werr == nil || !strings.Contains(werr.Error(), "TOML cursor desync") {
					t.Fatalf("walkTable(%q) error = %v, want a cursor desync error (row belongs in the desync-free table)", doc, werr)
				}
			}

			dir := t.TempDir()
			docPath := filepath.Join(dir, "doc.toml")
			mirrorPath := filepath.Join(dir, "mirror.toml")
			if err := os.WriteFile(docPath, []byte(row.doc), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
			if err := os.WriteFile(mirrorPath, []byte(row.mirror), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			posix := filepath.ToSlash(docPath)
			want := make([]string, 0, len(row.keyPaths))
			for i, kp := range row.keyPaths {
				segment := "channel_id"
				if row.segments != nil {
					segment = row.segments[i]
				}
				want = append(want, posix+": retired token 'channel_id' in TOML key path '"+kp+"' (segment '"+segment+"') (via sequence model)")
			}
			if got := scanTomlPrimary(docPath); !reflect.DeepEqual(got, want) {
				t.Fatalf("scanTomlPrimary(token doc) = %q, want %q", got, want)
			}
			if got := scanTomlPrimary(mirrorPath); len(got) != 0 {
				t.Fatalf("scanTomlPrimary(mirror doc) = %q, want no findings", got)
			}
		})
	}
}

// TestWalkDecodedTOML_MatchesWalkTable_DesyncFree is U1 scenario 2 (AC-2,
// characterization): on desync-free documents covering every decoded kind,
// walkDecodedTOML (the sorted fallback walk) reports exactly the same
// findings as walkTable (the cursor walk), compared as multisets.
func TestWalkDecodedTOML_MatchesWalkTable_DesyncFree(t *testing.T) {
	rows := []struct {
		name string
		doc  string
		min  int // minimum finding count, so a row can never pass vacuously
	}{
		{"scalars", "channel_id = 's'\nteam_id = 1\nipc_name = 1.5\nhost_cli = true\nslack = 1979-05-27T07:32:00Z\nk1 = 1979-05-27T07:32:00\nk2 = 1979-05-27\nk3 = 07:32:00\nplain = 'x'\n", 5},
		{"inline-table", "cfg = { channel_id = 1, inner = { team_id = 2 }, plain = 3 }\n", 2},
		{"array-of-inline-tables", "arr = [ { channel_id = 1, y = 2 }, { y = 3, team_id = 4 }, { channel_id = 5 } ]\n", 3},
		{"array-of-arrays", "arr = [ [ 1, 2 ], [ { channel_id = 1 } ], [ [ { team_id = 2 } ] ] ]\nchannel = [ 'x' ]\n", 2},
		{"array-of-tables", "[[t]]\nchannel_id = 1\n[[t.sub]]\nteam_id = 2\n[[t]]\nchannel_id = 3\nk = 4\n", 3},
		{"dotted-keys", "a.b.channel_id = 1\nchannel.id = 2\nsocket.mode.x = 3\n", 3},
		{"composed-header-path", "[channel]\nid = 1\n[team]\nids = 2\n", 2},
		{"ancestor-reopen", "[a.b]\nx = 1\n[a]\nchannel_id = 2\n", 1},
		{"array-of-tables-reopen", "[[t]]\nk = 1\n[t.sub]\nchannel_id = 1\n[[t]]\nm = 2\nteam_id = 3\n", 2},
		{"mixed-scalar-and-table-array", "arr = [ 1, 'two', { channel_id = 3 }, [ 4 ] ]\n", 1},
	}
	for _, row := range rows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			var data map[string]interface{}
			meta, err := toml.Decode(row.doc, &data)
			if err != nil {
				t.Fatalf("toml.Decode: %v", err)
			}
			var cursorFindings, fallbackFindings []string
			if err := walkTable("p", nil, data, &tomlCursor{keys: meta.Keys()}, &cursorFindings); err != nil {
				t.Fatalf("walkTable: %v (row must be desync-free)", err)
			}
			if err := walkDecodedTOML("p", nil, data, &fallbackFindings); err != nil {
				t.Fatalf("walkDecodedTOML: %v", err)
			}
			if len(cursorFindings) < row.min {
				t.Fatalf("walkTable findings = %q, want at least %d (row is vacuous)", cursorFindings, row.min)
			}
			if !sameFindingMultiset(cursorFindings, fallbackFindings) {
				t.Fatalf("multiset mismatch:\n walkTable       = %q\n walkDecodedTOML = %q", cursorFindings, fallbackFindings)
			}
		})
	}
}

// TestWalkDecodedTOML_UnexpectedType_FailsClosed is U1 scenario 3 (AC-3):
// a decoded value outside BurntSushi/toml v1.6.0's decoded type set must
// make the fallback walk fail closed with the "unexpected decoded type"
// error rather than being silently skipped.
func TestWalkDecodedTOML_UnexpectedType_FailsClosed(t *testing.T) {
	rows := []struct {
		name  string
		value map[string]interface{}
	}{
		{"go-int", map[string]interface{}{"k": int(1)}},
		{"string-slice", map[string]interface{}{"k": []string{"x"}}},
	}
	for _, row := range rows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			var findings []string
			err := walkDecodedTOML("p", nil, row.value, &findings)
			if err == nil || !strings.Contains(err.Error(), "unexpected decoded type") {
				t.Fatalf("walkDecodedTOML error = %v, want an 'unexpected decoded type' error", err)
			}
		})
	}
}

// TestScanTomlPrimaryWith_CompletenessOracle is U2 (AC-1..AC-3): on the
// cursor walk's success path the fallback walk runs as a completeness
// oracle, so a cursor walk that drops or invents a finding, or a fallback
// that fails, fails closed instead of returning the cursor findings.
func TestScanTomlPrimaryWith_CompletenessOracle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "two.toml")
	if err := os.WriteFile(path, []byte("channel_id = 1\nteam_id = 2\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	posix := filepath.ToSlash(path)
	if got := scanTomlPrimary(path); len(got) != 2 {
		t.Fatalf("precondition: scanTomlPrimary = %q, want 2 findings", got)
	}
	fallback := func(p string, prefix []string, v map[string]interface{}, _ *tomlCursor, f *[]string) error {
		return walkDecodedTOML(p, prefix, v, f)
	}
	completeness := []string{posix + ": TOML completeness check failed: cursor walk omitted or invented findings (fail-closed)"}

	t.Run("walker-drops-a-finding", func(t *testing.T) {
		drop := func(p string, prefix []string, v map[string]interface{}, c *tomlCursor, f *[]string) error {
			err := walkTable(p, prefix, v, c, f)
			if len(*f) > 0 {
				*f = (*f)[:len(*f)-1]
			}
			return err
		}
		if got := scanTomlPrimaryWith(path, drop, fallback); !reflect.DeepEqual(got, completeness) {
			t.Fatalf("scanTomlPrimaryWith(drop) = %q, want %q", got, completeness)
		}
	})
	t.Run("walker-invents-a-finding", func(t *testing.T) {
		invent := func(p string, prefix []string, v map[string]interface{}, c *tomlCursor, f *[]string) error {
			err := walkTable(p, prefix, v, c, f)
			*f = append(*f, p+": retired token 'acp' in TOML key path 'acp' (segment 'acp') (via sequence model)")
			return err
		}
		if got := scanTomlPrimaryWith(path, invent, fallback); !reflect.DeepEqual(got, completeness) {
			t.Fatalf("scanTomlPrimaryWith(invent) = %q, want %q", got, completeness)
		}
	})
	t.Run("fallback-errors", func(t *testing.T) {
		failing := func(string, []string, map[string]interface{}, *tomlCursor, *[]string) error {
			return errors.New("injected fallback failure")
		}
		want := []string{posix + ": TOML parse error (fail-closed): injected fallback failure"}
		if got := scanTomlPrimaryWith(path, walkTable, failing); !reflect.DeepEqual(got, want) {
			t.Fatalf("scanTomlPrimaryWith(failing fallback) = %q, want %q", got, want)
		}
	})
	// Review-fix (040-S local review): the desync path's own fallback
	// failure must also fail closed with exactly the parse-error finding,
	// discarding the cursor walk's partial findings, and the fallback must
	// be handed a nil cursor (it is cursor-free by contract).
	t.Run("desync-and-fallback-errors", func(t *testing.T) {
		desync := func(p string, prefix []string, v map[string]interface{}, c *tomlCursor, f *[]string) error {
			_ = walkTable(p, prefix, v, c, f)
			return errors.New("retiredarch: TOML cursor desync: injected")
		}
		var gotCursor *tomlCursor
		calls := 0
		failing := func(_ string, _ []string, _ map[string]interface{}, c *tomlCursor, _ *[]string) error {
			calls++
			gotCursor = c
			return errors.New("injected fallback failure")
		}
		want := []string{posix + ": TOML parse error (fail-closed): injected fallback failure"}
		if got := scanTomlPrimaryWith(path, desync, failing); !reflect.DeepEqual(got, want) {
			t.Fatalf("scanTomlPrimaryWith(desync, failing fallback) = %q, want %q", got, want)
		}
		if calls != 1 || gotCursor != nil {
			t.Fatalf("fallback calls = %d, cursor = %v; want exactly 1 call with a nil cursor", calls, gotCursor)
		}
	})
}
