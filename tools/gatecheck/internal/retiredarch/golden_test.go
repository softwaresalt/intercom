package retiredarch

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot returns the repository top-level directory from this test
// package's own path
// (tools/gatecheck/internal/retiredarch), so tests can exercise the real
// fixture corpus and the real tracked internal/**, cmd/** tree without
// depending on the working directory `go test` happens to use.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return root
}

// retiredarchGolden mirrors tools/gatecheck/internal/retiredarch/testdata/
// retiredarch_golden.json, captured from scripts/lib/retired_arch.py at
// the M2 parent commit (see docs/plans/evidence/2026-09-28-gate-engine-go-
// migration/m2.md).
type retiredarchGolden struct {
	ForbiddenParts      map[string][]string      `json:"forbidden_parts"`
	ForbiddenPartsOrder []string                 `json:"forbidden_parts_order"`
	VocabWords          []string                 `json:"vocab_words"`
	RootPlaceholder     string                   `json:"root_placeholder"`
	PythonVersion       string                   `json:"python_version"`
	UnidataVersion      string                   `json:"unidata_version"`
	GoUnicodeVersion    string                   `json:"go_unicode_version"`
	GoFixtureFindings   []goFixtureFindingRow    `json:"go_fixture_findings"`
	TomlFixtureFindings []tomlFixtureFindingRow  `json:"toml_fixture_findings"`
	IdentGoldens        []identGoldenRow         `json:"ident_goldens"`
	InlineTomlGoldens   map[string]inlineTomlRow `json:"inline_toml_goldens"`
	StreamCaptures      map[string]streamRow     `json:"stream_captures"`
	RepoModeEvidence    repoModeEvidence         `json:"repo_mode_evidence"`
}

type goFixtureFindingRow struct {
	Name             string   `json:"name"`
	MaskedFindings   []string `json:"masked_findings"`
	UnmaskedFindings []string `json:"unmasked_findings"`
}

type tomlFixtureFindingRow struct {
	Name             string   `json:"name"`
	TomllibFindings  []string `json:"tomllib_findings"`
	FallbackFindings []string `json:"fallback_findings"`
	Agree            bool     `json:"agree"`
}

type identGoldenRow struct {
	Name                     string       `json:"name"`
	SplitIdentifier          []string     `json:"split_identifier"`
	SegmentWholeWhole        [][]string   `json:"segment_whole_whole"`
	SegmentWholePerComponent [][][]string `json:"segment_whole_per_component"`
}

type inlineTomlRow struct {
	Text             string   `json:"text"`
	RawBytes         []int    `json:"raw_bytes"`
	TomllibFindings  []string `json:"tomllib_findings"`
	FallbackFindings []string `json:"fallback_findings"`
	Agree            bool     `json:"agree"`
}

type streamRow struct {
	Cmd        []string          `json:"cmd"`
	Cwd        string            `json:"cwd"`
	Stdout     string            `json:"stdout"`
	Stderr     string            `json:"stderr"`
	ExitCode   int               `json:"exit_code"`
	Assertions []assertionRecord `json:"assertions"`
}

// assertionRecord mirrors one {name, outcome} entry captured by the M2-T1
// generator for a self-test/self-test-integrity stream (C-5: live-count
// assertions are compared by name+outcome only, never by their exact
// corpus-count text).
type assertionRecord struct {
	Name    string `json:"name"`
	Outcome string `json:"outcome"`
}

type repoModeEvidence struct {
	NormalCwd         streamRow `json:"normal_cwd"`
	SubdirectoryCwd   streamRow `json:"subdirectory_cwd"`
	AnotherRepoCwd    streamRow `json:"another_repo_cwd"`
	SelectedPathSet   []string  `json:"selected_path_set"`
	SelectedPathCount int       `json:"selected_path_count"`
}

func loadGolden(t *testing.T) retiredarchGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "retiredarch_golden.json"))
	if err != nil {
		t.Fatalf("read retiredarch_golden.json: %v", err)
	}
	var g retiredarchGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal retiredarch_golden.json: %v", err)
	}
	return g
}

// derootify replaces every occurrence of the golden's root placeholder
// token with root (posix-formatted), reversing the substitution the M2-T1
// generator applied so absolute-path-embedding findings (retired_arch.py's
// finding-format functions all embed path.as_posix() where path = root /
// rel_path, root ABSOLUTE) can be compared against a portable golden.
// derootify replaces every occurrence of placeholder (already including
// the \u0001...\u0001 wrapper bytes as captured verbatim in
// g.RootPlaceholder) with root.
func derootify(s, placeholder, root string) string {
	return strings.ReplaceAll(s, placeholder, root)
}

func derootifyAll(items []string, placeholder, root string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = derootify(s, placeholder, root)
	}
	return out
}
