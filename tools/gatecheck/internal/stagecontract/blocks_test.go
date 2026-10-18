package stagecontract

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func joinLines(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

func assertBlocks(t *testing.T, src []byte, want []Block) {
	t.Helper()
	got, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d blocks, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("block %d:\n got  %+v\n want %+v", i, got[i], want[i])
		}
	}
}

func TestParseHeadingsAndLists(t *testing.T) {
	cases := []struct {
		name string
		src  []byte
		want []Block
	}{
		{
			name: "heading path and list markers",
			src: joinLines(
				"# Title",
				"intro text",
				"## Sub",
				"- a",
				"  cont",
				"* b",
				"1. c",
				"### Deep",
				"2) d",
				"## Next",
				"- e",
			),
			want: []Block{
				{Kind: KindHeading, Line: 1, Level: 1, Text: "Title", HeadingPath: []string{"Title"}},
				{Kind: KindParagraph, Line: 2, Text: "intro text", HeadingPath: []string{"Title"}},
				{Kind: KindHeading, Line: 3, Level: 2, Text: "Sub", HeadingPath: []string{"Title", "Sub"}},
				{Kind: KindListItem, Line: 4, Text: "a cont", HeadingPath: []string{"Title", "Sub"}},
				{Kind: KindListItem, Line: 6, Text: "b", HeadingPath: []string{"Title", "Sub"}},
				{Kind: KindListItem, Line: 7, Text: "c", HeadingPath: []string{"Title", "Sub"}},
				{Kind: KindHeading, Line: 8, Level: 3, Text: "Deep", HeadingPath: []string{"Title", "Sub", "Deep"}},
				{Kind: KindListItem, Line: 9, Text: "d", HeadingPath: []string{"Title", "Sub", "Deep"}},
				{Kind: KindHeading, Line: 10, Level: 2, Text: "Next", HeadingPath: []string{"Title", "Next"}},
				{Kind: KindListItem, Line: 11, Text: "e", HeadingPath: []string{"Title", "Next"}},
			},
		},
		{
			name: "blocks before any heading have no path",
			src:  joinLines("- first", "", "# H"),
			want: []Block{
				{Kind: KindListItem, Line: 1, Text: "first"},
				{Kind: KindHeading, Line: 3, Level: 1, Text: "H", HeadingPath: []string{"H"}},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { assertBlocks(t, tc.src, tc.want) })
	}
}

func TestParseTables(t *testing.T) {
	cases := []struct {
		name string
		src  []byte
		want []Block
	}{
		{
			name: "header separator and rows",
			src: joinLines(
				"## T",
				"| Col A | Col B |",
				"|-------|:-----:|",
				"| x | y |",
				"| z |  |",
			),
			want: []Block{
				{Kind: KindHeading, Line: 1, Level: 2, Text: "T", HeadingPath: []string{"T"}},
				{Kind: KindTableRow, Line: 2, Text: "| Col A | Col B |", HeadingPath: []string{"T"}, Cells: []string{"Col A", "Col B"}},
				{Kind: KindTableRow, Line: 4, Text: "| x | y |", HeadingPath: []string{"T"}, Cells: []string{"x", "y"}},
				{Kind: KindTableRow, Line: 5, Text: "| z |  |", HeadingPath: []string{"T"}, Cells: []string{"z", ""}},
			},
		},
		{
			name: "table ends at prose",
			src:  joinLines("| a | b |", "after"),
			want: []Block{
				{Kind: KindTableRow, Line: 1, Text: "| a | b |", Cells: []string{"a", "b"}},
				{Kind: KindParagraph, Line: 2, Text: "after"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { assertBlocks(t, tc.src, tc.want) })
	}
}

func TestParseFencedQuoteHistorical(t *testing.T) {
	src := joinLines(
		"# Doc",
		"> quoted line",
		"> more",
		"text",
		"~~~text",
		"# not heading",
		"- not list",
		"~~~",
		"## Historical notes",
		"- old item",
		"### Sub",
		"| a | b |",
		"## Example: usage",
		"```go",
		"x",
		"```",
		"## Live",
		"- ok",
	)
	hist := []string{"Doc", "Historical notes"}
	want := []Block{
		{Kind: KindHeading, Line: 1, Level: 1, Text: "Doc", HeadingPath: []string{"Doc"}},
		{Kind: KindBlockquote, Line: 2, Text: "quoted line\nmore", HeadingPath: []string{"Doc"}, Lines: []string{"quoted line", "more"}, Quoted: true},
		{Kind: KindParagraph, Line: 4, Text: "text", HeadingPath: []string{"Doc"}},
		{Kind: KindFenced, Line: 5, Text: "# not heading\n- not list", HeadingPath: []string{"Doc"}, Lines: []string{"# not heading", "- not list"}, Lang: "text", Quoted: true},
		{Kind: KindHeading, Line: 9, Level: 2, Text: "Historical notes", HeadingPath: hist, Historical: true},
		{Kind: KindListItem, Line: 10, Text: "old item", HeadingPath: hist, Historical: true},
		{Kind: KindHeading, Line: 11, Level: 3, Text: "Sub", HeadingPath: []string{"Doc", "Historical notes", "Sub"}, Historical: true},
		{Kind: KindTableRow, Line: 12, Text: "| a | b |", HeadingPath: []string{"Doc", "Historical notes", "Sub"}, Cells: []string{"a", "b"}, Historical: true},
		{Kind: KindHeading, Line: 13, Level: 2, Text: "Example: usage", HeadingPath: []string{"Doc", "Example: usage"}, Historical: true},
		{Kind: KindFenced, Line: 14, Text: "x", HeadingPath: []string{"Doc", "Example: usage"}, Lines: []string{"x"}, Lang: "go", Quoted: true, Historical: true},
		{Kind: KindHeading, Line: 17, Level: 2, Text: "Live", HeadingPath: []string{"Doc", "Live"}},
		{Kind: KindListItem, Line: 18, Text: "ok", HeadingPath: []string{"Doc", "Live"}},
	}
	assertBlocks(t, src, want)

	if _, err := Parse(joinLines("```text", "never closed")); err == nil {
		t.Error("unterminated fence: want error")
	}
}

func TestParseRealContractFiles(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	files := []string{
		filepath.Join(root, ".github", "policies", "workflow-policies.md"),
		filepath.Join(root, ".github", "agents", "_stage.agent.md"),
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read contract file: %v", err)
			}
			blocks, err := Parse(src)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			counts := map[BlockKind]int{}
			for _, b := range blocks {
				counts[b.Kind]++
			}
			if counts[KindHeading] == 0 {
				t.Errorf("no headings extracted from %s", f)
			}
		})
	}
}
