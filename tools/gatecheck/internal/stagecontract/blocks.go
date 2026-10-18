package stagecontract

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

// Parse extracts the markdown block structure of src. It performs no
// evaluation of the content and returns an error only for malformed
// structure such as an unterminated fenced block.
func Parse(src []byte) ([]Block, error) {
	panic("not implemented: Parse markdown block extraction")
}
