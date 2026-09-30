package retiredarch

import (
	"reflect"
	"testing"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

func TestVocabWords_MatchesGoldenSet(t *testing.T) {
	g := loadGolden(t)
	// Per-run tie-break order among words of equal length is a documented
	// non-issue (see vocabWords' doc comment): compare as a MULTISET
	// (sorted-by-length groups equal), not a literal order, against the
	// golden capture, which itself reflects only ONE of many valid
	// Python tie-break orderings.
	if len(vocabWords) != len(g.VocabWords) {
		t.Fatalf("vocabWords has %d entries, golden has %d: %v vs %v", len(vocabWords), len(g.VocabWords), vocabWords, g.VocabWords)
	}
	gotSet := map[string]bool{}
	for _, w := range vocabWords {
		gotSet[w] = true
	}
	for _, w := range g.VocabWords {
		if !gotSet[w] {
			t.Fatalf("vocabWords missing golden word %q; got %v", w, vocabWords)
		}
	}
	// Length-descending grouping must match exactly (ties aside).
	for i := 1; i < len(vocabWords); i++ {
		if len(vocabWords[i]) > len(vocabWords[i-1]) {
			t.Fatalf("vocabWords not sorted by length descending at index %d: %v", i, vocabWords)
		}
	}
}

func TestForbiddenParts_MatchesGoldenOrderAndContent(t *testing.T) {
	g := loadGolden(t)
	if len(forbiddenParts) != len(g.ForbiddenPartsOrder) {
		t.Fatalf("forbiddenParts has %d entries, golden order has %d", len(forbiddenParts), len(g.ForbiddenPartsOrder))
	}
	for i, fp := range forbiddenParts {
		if fp.token != g.ForbiddenPartsOrder[i] {
			t.Fatalf("forbiddenParts[%d].token = %q, golden order says %q", i, fp.token, g.ForbiddenPartsOrder[i])
		}
		want, ok := g.ForbiddenParts[fp.token]
		if !ok {
			t.Fatalf("golden forbidden_parts missing token %q", fp.token)
		}
		if !reflect.DeepEqual(fp.want, want) {
			t.Fatalf("forbiddenParts[%q].want = %v, golden = %v", fp.token, fp.want, want)
		}
	}
}

func TestSplitIdentifier_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	for _, row := range g.IdentGoldens {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			got := splitIdentifier(row.Name)
			if !reflect.DeepEqual(got, row.SplitIdentifier) {
				t.Fatalf("splitIdentifier(%q) = %v, want %v", row.Name, got, row.SplitIdentifier)
			}
		})
	}
}

func TestSegmentWhole_MatchesGolden_Whole(t *testing.T) {
	g := loadGolden(t)
	for _, row := range g.IdentGoldens {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			whole := lowerWholeForTest(row.Name)
			got := segmentWhole(whole)
			want := row.SegmentWholeWhole
			if len(got) == 0 && len(want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("segmentWhole(%q) = %v, want %v", whole, got, want)
			}
		})
	}
}

func TestSegmentWhole_MatchesGolden_PerComponent(t *testing.T) {
	g := loadGolden(t)
	for _, row := range g.IdentGoldens {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			parts := splitIdentifier(row.Name)
			if len(parts) != len(row.SegmentWholePerComponent) {
				t.Fatalf("splitIdentifier(%q) produced %d parts, golden has %d segment_whole_per_component entries", row.Name, len(parts), len(row.SegmentWholePerComponent))
			}
			for i, component := range parts {
				got := segmentWhole(component)
				want := row.SegmentWholePerComponent[i]
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("segmentWhole(%q) (component %d of %q) = %v, want %v", component, i, row.Name, got, want)
				}
			}
		})
	}
}

// lowerWholeForTest reproduces how the M2-T1 generator computed
// segment_whole_whole: word.lower() applied to the whole identifier name
// (Python str.lower(), i.e. pysem.Lower), NOT splitIdentifier's
// underscore/camel-aware split.
func lowerWholeForTest(name string) string {
	return pysem.Lower(name)
}
