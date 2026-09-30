package gomask

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type maskerGoldenEntry struct {
	Name   string `json:"name"`
	Input  string `json:"input"`
	Masked string `json:"masked"`
}

func loadMaskerGolden(t *testing.T) []maskerGoldenEntry {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "masker_golden.json"))
	if err != nil {
		t.Fatalf("read masker_golden.json: %v", err)
	}
	var entries []maskerGoldenEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("unmarshal masker_golden.json: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("masker_golden.json contains no entries")
	}
	return entries
}

func TestMaskGoNonCode_Golden(t *testing.T) {
	entries := loadMaskerGolden(t)
	for _, e := range entries {
		e := e
		t.Run(e.Name, func(t *testing.T) {
			got := MaskGoNonCode(e.Input)
			if got != e.Masked {
				t.Fatalf("MaskGoNonCode(%q) = %q, want %q", e.Input, got, e.Masked)
			}
			if len([]rune(got)) != len([]rune(e.Input)) {
				t.Fatalf("MaskGoNonCode(%q): rune length %d != input rune length %d",
					e.Input, len([]rune(got)), len([]rune(e.Input)))
			}
		})
	}
}

// TestMaskGoNonCode_MultilineTagShapedRawStringPinned pins the D-4
// characterization: a raw-string literal whose content is two struct-tag
// pairs separated by an embedded literal newline (rather than a run of
// non-newline whitespace) is still treated as a single, unmasked struct
// tag, because the Unicode-aware \s class (and CPython's non-MULTILINE
// regex anchors) the ported struct_tag_re relies on both include '\n'.
// This is intentional, pinned legacy behavior (stash C312BD4C is out of
// scope for this port -- see plan D-4) and must not be "fixed" here.
func TestMaskGoNonCode_MultilineTagShapedRawStringPinned(t *testing.T) {
	const name = "multiline-tag-shaped-raw-string"
	entries := loadMaskerGolden(t)
	for _, e := range entries {
		if e.Name == name {
			got := MaskGoNonCode(e.Input)
			if got != e.Masked {
				t.Fatalf("pinned characterization %s regressed: got %q, want %q", name, got, e.Masked)
			}
			return
		}
	}
	t.Fatalf("golden entry %q not found in masker_golden.json", name)
}

func TestIsStructTag(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"single benign pair", `json:"team_name"`, true},
		{"leading and trailing space", ` json:"a" `, true},
		{"two pairs space separated", `x:"a" y:"b"`, true},
		{"two pairs nbsp separated", "x:\"a\"\u00a0y:\"b\"", true},
		{"two pairs newline separated", "json:\"a\"\njson:\"b\"", true},
		{"adjacent pairs no space", `x:"a"y:"b"`, false},
		{"digit leading key", `1x:"a"`, false},
		{"empty content", ``, false},
		{"non-tag prose", `SELECT * FROM t WHERE channel_id = ?`, false},
		{"value contains control char", "x:\"a\x1cb\"", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isStructTag([]rune(tc.content))
			if got != tc.want {
				t.Fatalf("isStructTag(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}
