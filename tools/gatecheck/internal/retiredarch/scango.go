// This file (scango.go) ports scan_go from scripts/lib/retired_arch.py.
package retiredarch

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// goIdentifierRe matches an ASCII Go-identifier-shaped token
// ([A-Za-z_][A-Za-z0-9_]*), WITHOUT Python's `\b` boundary assertions
// built into the pattern itself (go_identifier_re =
// re.compile(r'\b[A-Za-z_][A-Za-z0-9_]*\b') in scripts/lib/retired_arch.py
// -- RE2/Go's regexp package has no zero-width-lookaround primitive for
// this). Every candidate match found by this pattern is instead
// re-validated by checking pysem.WordBoundary (Unicode-\w-aware, matching
// Python's own `\b` under a str pattern) at BOTH the match's start and end
// byte offsets.
//
// This is mathematically equivalent to Python's
// `\b[A-Za-z_][A-Za-z0-9_]*\b`: every INTERIOR position of a maximal
// [A-Za-z0-9_] run is always word-word on both sides (so `\b` can never
// hold there), and the character class's first-char restriction
// ([A-Za-z_], excluding digits) means a match can only ever be attempted
// starting at the run's true beginning or at some position preceded by a
// non-word-class char that is nonetheless itself a Unicode word char (in
// which case the start-boundary check below correctly rejects it, exactly
// as Python's `\b` would) -- so the only positions where `\b` could ever
// hold are the two offsets this recheck inspects. This also correctly
// yields NO match at all when the class-restricted match abuts an
// adjacent Unicode word rune (e.g. "café123": Go's regexp finds "caf" as
// its only candidate here since digits can't start a new match and 'é'
// isn't in the class; the end-boundary check at the offset just after
// "caf" sees 'é' as IsWord and rejects it -- matching Python's own
// zero-match behavior for that input exactly).
var goIdentifierRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// scanGo ports scan_go. mask=true (Python's default) is the real,
// production self-test path (gomask.MaskGoNonCode applied first);
// mask=false is the go-differential fixture suite's masking-bypass seam,
// scanning raw (unmasked) source text directly.
//
// Line numbering uses pysem.SplitLines (masked.splitlines() in Python),
// not strings.Split(text, "\n"): Python's str.splitlines() also breaks on
// \r, \v, \f, \x1c-\x1e, NEL and U+2028/U+2029, any of which the
// masked/raw Go source text could in principle contain inside a
// string/comment/rune literal now visible post-masking, and this must
// match Python's own line numbering exactly.
func scanGo(path string, mask bool) ([]string, error) {
	text, err := pysem.ReadText(path)
	if err != nil {
		return nil, err
	}
	masked := text
	if mask {
		masked = gomask.MaskGoNonCode(text)
	}
	posixPath := filepath.ToSlash(path)

	var findings []string
	for i, line := range pysem.SplitLines(masked) {
		lineNo := i + 1
		for _, loc := range goIdentifierRe.FindAllStringIndex(line, -1) {
			start, end := loc[0], loc[1]
			if !pysem.WordBoundary(line, start) || !pysem.WordBoundary(line, end) {
				continue
			}
			ident := line[start:end]
			token, model, ok := matchesForbiddenParts(splitIdentifier(ident))
			if !ok {
				continue
			}
			findings = append(findings, fmt.Sprintf(
				"%s:%d: retired token %s in Go identifier %s (via %s model)",
				posixPath, lineNo, pysem.Repr(token), pysem.Repr(ident), model,
			))
		}
	}
	return findings, nil
}
