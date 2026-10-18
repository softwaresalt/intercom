package stagecontract

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// BlockKind names the markdown structure a Block was extracted from.
type BlockKind string

// Block kinds produced by Parse.
const (
	KindHeading    BlockKind = "heading"
	KindTableRow   BlockKind = "table_row"
	KindListItem   BlockKind = "list_item"
	KindFenced     BlockKind = "fenced"
	KindBlockquote BlockKind = "blockquote"
	KindParagraph  BlockKind = "paragraph"
)

// Block is one structural element of a markdown document. It carries no verdict.
type Block struct {
	Kind BlockKind
	// Line is the 1-based source line on which the block starts.
	Line int
	// Level is the heading level (1-6); zero for other kinds.
	Level int
	// Text is the block text: heading text, list item text, the trimmed table
	// row, or the joined lines of a paragraph, blockquote or fenced block.
	Text string
	// HeadingPath lists the enclosing headings, outermost first. A heading
	// block includes itself.
	HeadingPath []string
	// Cells holds table row cells with surrounding whitespace removed.
	Cells []string
	// Lines holds the content lines of a fenced block or blockquote.
	Lines []string
	// Lang is the first word of a fenced block info string.
	Lang string
	// Quoted is true for blockquotes and fenced blocks.
	Quoted bool
	// Historical is true inside a section whose heading path names it
	// historical or example.
	Historical bool
}

var (
	headingRe    = regexp.MustCompile(`^ {0,3}(#{1,6})[ \t]+(.*?)(?:[ \t]+#+)?[ \t]*$`)
	listItemRe   = regexp.MustCompile(`^[ \t]*(?:[-*+]|\d+[.)])[ \t]+(.*)$`)
	fenceRe      = regexp.MustCompile("^[ \\t]*(`{3,}|~{3,})[ \\t]*(\\S*)")
	breakRe      = regexp.MustCompile(`^ {0,3}([-*_][ \t]*){3,}$`)
	separatorRe  = regexp.MustCompile(`^:?-+:?$`)
	historicalRe = regexp.MustCompile(`(?i)\b(historical|examples?)\b`)
)

// Parse extracts the markdown block structure of src. It performs no
// evaluation of the content and returns an error only for malformed
// structure such as an unterminated fenced block.
func Parse(src []byte) ([]Block, error) {
	lines := strings.Split(string(src), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, "\r")
	}

	p := &parser{}
	i := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for j := 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "---" {
				i = j + 1
				break
			}
		}
	}

	for ; i < len(lines); i++ {
		line := lines[i]
		no := i + 1
		trimmed := strings.TrimSpace(line)

		if m := fenceRe.FindStringSubmatch(line); m != nil {
			next, err := p.fenced(lines, i, m)
			if err != nil {
				return nil, err
			}
			i = next
			continue
		}
		switch {
		case trimmed == "":
			p.last = nil
		case headingRe.MatchString(line):
			m := headingRe.FindStringSubmatch(line)
			p.heading(no, len(m[1]), m[2])
		case strings.HasPrefix(strings.TrimLeft(line, " "), ">") && len(line)-len(strings.TrimLeft(line, " ")) < 4:
			var quoted []string
			start := no
			for ; i < len(lines); i++ {
				l := strings.TrimLeft(lines[i], " ")
				if !strings.HasPrefix(l, ">") {
					break
				}
				quoted = append(quoted, strings.TrimPrefix(strings.TrimPrefix(l, ">"), " "))
			}
			i--
			b := p.block(KindBlockquote, start, strings.Join(quoted, "\n"))
			b.Lines = quoted
			b.Quoted = true
			p.add(b)
			p.last = nil
		case strings.HasPrefix(trimmed, "|"):
			b := p.block(KindTableRow, no, trimmed)
			b.Cells = splitCells(trimmed)
			if isSeparator(b.Cells) {
				p.last = nil
				continue
			}
			p.add(b)
			p.last = nil
		case breakRe.MatchString(line):
			p.last = nil
		case listItemRe.MatchString(line):
			m := listItemRe.FindStringSubmatch(line)
			p.add(p.block(KindListItem, no, strings.TrimSpace(m[1])))
			p.last = &p.blocks[len(p.blocks)-1]
		default:
			if p.last != nil && (p.last.Kind == KindParagraph || p.last.Kind == KindListItem) {
				p.last.Text += " " + trimmed
				continue
			}
			p.add(p.block(KindParagraph, no, trimmed))
			p.last = &p.blocks[len(p.blocks)-1]
		}
	}
	return p.blocks, nil
}

type parser struct {
	blocks []Block
	path   []string
	levels []int
	hist   bool
	last   *Block
}

func (p *parser) block(kind BlockKind, line int, text string) Block {
	return Block{
		Kind:        kind,
		Line:        line,
		Text:        text,
		HeadingPath: slices.Clone(p.path),
		Historical:  p.hist,
	}
}

// add appends b; callers reset p.last themselves because the slice may move.
func (p *parser) add(b Block) {
	p.blocks = append(p.blocks, b)
}

func (p *parser) heading(line, level int, text string) {
	text = strings.TrimSpace(text)
	for len(p.levels) > 0 && p.levels[len(p.levels)-1] >= level {
		p.levels = p.levels[:len(p.levels)-1]
		p.path = p.path[:len(p.path)-1]
	}
	p.levels = append(p.levels, level)
	p.path = append(p.path, text)
	p.hist = false
	for _, h := range p.path {
		if historicalRe.MatchString(h) {
			p.hist = true
		}
	}
	b := p.block(KindHeading, line, text)
	b.Level = level
	p.add(b)
	p.last = nil
}

// fenced consumes a fenced block starting at index start and returns the
// index of its closing fence line.
func (p *parser) fenced(lines []string, start int, m []string) (int, error) {
	marker := m[1]
	var body []string
	for j := start + 1; j < len(lines); j++ {
		t := strings.TrimSpace(lines[j])
		if strings.HasPrefix(t, marker[:1]) && strings.Trim(t, marker[:1]) == "" && len(t) >= len(marker) {
			b := p.block(KindFenced, start+1, strings.Join(body, "\n"))
			b.Lines = body
			b.Lang = m[2]
			b.Quoted = true
			p.add(b)
			p.last = nil
			return j, nil
		}
		body = append(body, lines[j])
	}
	return 0, fmt.Errorf("unterminated fenced block starting at line %d", start+1)
}

func splitCells(row string) []string {
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	var cells []string
	var cur strings.Builder
	inSpan := false
	for _, r := range row {
		switch {
		case r == '`':
			inSpan = !inSpan
			cur.WriteRune(r)
		case r == '|' && !inSpan:
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	return append(cells, strings.TrimSpace(cur.String()))
}

func isSeparator(cells []string) bool {
	for _, c := range cells {
		if !separatorRe.MatchString(c) {
			return false
		}
	}
	return true
}
