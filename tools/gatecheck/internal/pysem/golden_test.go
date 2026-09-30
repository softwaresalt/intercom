package pysem

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// runeCase is one entry of the per-rune-classification golden sections
// (is_space, is_word, is_digit, is_decimal, is_upper, is_lower).
type runeCase struct {
	CP     int    `json:"cp"`
	Input  string `json:"input"`
	Expect bool   `json:"expect"`
}

// lowerCase is one entry of the "lower" golden section.
type lowerCase struct {
	Input  string `json:"input"`
	Expect string `json:"expect"`
}

// stripCase is one entry of the "strip" golden section.
type stripCase struct {
	Input  string `json:"input"`
	Expect string `json:"expect"`
}

// splitLinesCase is one entry of the "split_lines" golden section.
type splitLinesCase struct {
	Input  string   `json:"input"`
	Expect []string `json:"expect"`
}

// reprCase is one entry of the "repr" golden section.
type reprCase struct {
	Input  string `json:"input"`
	Expect string `json:"expect"`
}

// offsetCase is one entry of the word_boundary, preceded_by_word_or_dot, and
// followed_by_word golden sections: text plus a byte offset into it, and
// the expected boolean at that offset.
type offsetCase struct {
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Expect bool   `json:"expect"`
}

// readTextCase is one entry of the "read_text" golden section. Expect is a
// pointer because the invalid-UTF-8 case has no defined translated string
// (Python's read_text raised UnicodeDecodeError; the JSON records
// expect: null).
type readTextCase struct {
	Name         string  `json:"name"`
	RawUTF8Bytes []int   `json:"raw_utf8_bytes"`
	Expect       *string `json:"expect"`
	InvalidUTF8  bool    `json:"invalid_utf8"`
}

// pysemGolden mirrors testdata/pysem_golden.json in full so every
// register_<name>_test.go / *_test.go in this package can load exactly the
// sections it needs from one shared, single-parse fixture.
type pysemGolden struct {
	PythonVersion       string           `json:"python_version"`
	UnidataVersion      string           `json:"unidata_version"`
	GoUnicodeVersion    string           `json:"go_unicode_version"`
	IsSpace             []runeCase       `json:"is_space"`
	IsWord              []runeCase       `json:"is_word"`
	IsDigit             []runeCase       `json:"is_digit"`
	IsDecimal           []runeCase       `json:"is_decimal"`
	IsUpper             []runeCase       `json:"is_upper"`
	IsLower             []runeCase       `json:"is_lower"`
	Lower               []lowerCase      `json:"lower"`
	Strip               []stripCase      `json:"strip"`
	SplitLines          []splitLinesCase `json:"split_lines"`
	Repr                []reprCase       `json:"repr"`
	WordBoundary        []offsetCase     `json:"word_boundary"`
	PrecededByWordOrDot []offsetCase     `json:"preceded_by_word_or_dot"`
	FollowedByWord      []offsetCase     `json:"followed_by_word"`
	ReadText            []readTextCase   `json:"read_text"`
}

// loadGolden reads and parses testdata/pysem_golden.json once per calling
// test. It fails the test (rather than returning an error) since every
// caller treats a missing/malformed golden file as fatal.
func loadGolden(t *testing.T) pysemGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "pysem_golden.json"))
	if err != nil {
		t.Fatalf("read pysem_golden.json: %v", err)
	}
	var g pysemGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal pysem_golden.json: %v", err)
	}
	return g
}

// runeOf returns the first (and, for every golden rune-classification
// entry, only) rune of c.Input.
func runeOf(t *testing.T, c runeCase) rune {
	t.Helper()
	rs := []rune(c.Input)
	if len(rs) != 1 {
		t.Fatalf("rune case %+v: input is not exactly one rune (%d)", c, len(rs))
	}
	return rs[0]
}
