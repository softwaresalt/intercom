// Package gomask reimplements the repository's canonical Go non-code masker
// (scripts/lib/gomask.py: mask_go_non_code / struct_tag_re) on top of
// tools/gatecheck/internal/pysem, so the write-path and retired-architecture
// gates can run without a Python interpreter.
//
// MaskGoNonCode blanks out comment, interpreted-string, rune, and raw-string
// literal contents while preserving total rune length and line structure, so
// a token appearing only in non-code text is never mistaken for code and
// finding line numbers stay aligned with the source file. A raw-string
// literal whose ENTIRE content matches the struct-tag grammar is left
// visible (struct tags carry config keys); an unterminated raw string at EOF
// fails closed (masked).
//
// The struct-tag matcher is a hand-rolled rune-level parser rather than a
// Go regexp: RE2's \s and \w character classes are ASCII-only by default
// and do not match CPython's Unicode-aware \s/\w (the Python source this
// package ports uses Unicode \s/\w), so the parser below calls
// pysem.IsSpace / pysem.IsWord directly instead.
package gomask

import "github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"

// maskerState is the mask_go_non_code() lexer state.
type maskerState int

const (
	stateCode maskerState = iota
	stateLineComment
	stateBlockComment
	stateString
	stateRawString
	stateRuneLit
)

// MaskGoNonCode blanks out Go comment, interpreted-string, rune, and
// raw-string literal contents (except a raw-string literal whose entire
// content is a Go struct tag), preserving rune count and line structure.
// This is a byte-identical (rune-identical) port of Python's
// mask_go_non_code(); see scripts/lib/gomask.py for the reference
// implementation and rationale.
func MaskGoNonCode(text string) string {
	runes := []rune(text)
	n := len(runes)
	out := make([]rune, 0, n)
	var rawBuf []rune

	state := stateCode
	i := 0
	for i < n {
		ch := runes[i]
		var nxt rune = -1
		if i+1 < n {
			nxt = runes[i+1]
		}

		switch state {
		case stateCode:
			switch {
			case ch == '/' && nxt == '/':
				out = append(out, ' ', ' ')
				i += 2
				state = stateLineComment
				continue
			case ch == '/' && nxt == '*':
				out = append(out, ' ', ' ')
				i += 2
				state = stateBlockComment
				continue
			case ch == '"':
				out = append(out, ' ')
				i++
				state = stateString
				continue
			case ch == '`':
				out = append(out, ' ')
				i++
				state = stateRawString
				rawBuf = rawBuf[:0]
				continue
			case ch == '\'':
				out = append(out, ' ')
				i++
				state = stateRuneLit
				continue
			default:
				out = append(out, ch)
				i++
				continue
			}

		case stateLineComment:
			if ch == '\n' {
				out = append(out, '\n')
				state = stateCode
			} else {
				out = append(out, ' ')
			}
			i++
			continue

		case stateBlockComment:
			if ch == '*' && nxt == '/' {
				out = append(out, ' ', ' ')
				i += 2
				state = stateCode
			} else {
				if ch == '\n' {
					out = append(out, '\n')
				} else {
					out = append(out, ' ')
				}
				i++
			}
			continue

		case stateString:
			if ch == '\\' && nxt != -1 {
				out = append(out, ' ', ' ')
				i += 2
				continue
			}
			if ch == '\n' {
				out = append(out, '\n')
			} else {
				out = append(out, ' ')
			}
			i++
			if ch == '"' {
				state = stateCode
			}
			continue

		case stateRawString:
			if ch == '`' {
				// Decide whole-content struct-tag visibility only now that
				// the ENTIRE raw-string content is known -- this is why the
				// content must be buffered rather than masked rune-by-rune
				// as it streams past. Only a literal whose entire content
				// matches the tag grammar is unmasked; everything else
				// (multiline strings, SQL/template backtick literals, etc.)
				// keeps the fully-masked behavior.
				if isStructTag(rawBuf) {
					out = append(out, rawBuf...)
				} else {
					for _, c := range rawBuf {
						if c == '\n' {
							out = append(out, '\n')
						} else {
							out = append(out, ' ')
						}
					}
				}
				out = append(out, ' ')
				i++
				state = stateCode
				rawBuf = rawBuf[:0]
				continue
			}
			rawBuf = append(rawBuf, ch)
			i++
			continue

		case stateRuneLit:
			if ch == '\\' && nxt != -1 {
				out = append(out, ' ', ' ')
				i += 2
				continue
			}
			if ch == '\n' {
				out = append(out, '\n')
			} else {
				out = append(out, ' ')
			}
			i++
			if ch == '\'' {
				state = stateCode
			}
			continue
		}
	}

	if state == stateRawString && len(rawBuf) > 0 {
		// Unterminated raw string at EOF: fail closed by masking the
		// buffered content -- never unmask a literal whose content could
		// not be fully determined to be (or not be) a complete struct tag.
		for _, c := range rawBuf {
			if c == '\n' {
				out = append(out, '\n')
			} else {
				out = append(out, ' ')
			}
		}
	}

	return string(out)
}

// isStructTag reports whether content is (in its entirety) a Go struct tag:
//
//	^\s*[A-Za-z_]\w*:"[^"]*"(?:\s+[A-Za-z_]\w*:"[^"]*")*\s*$
//
// where \s and \w are CPython's Unicode-aware classes (pysem.IsSpace /
// pysem.IsWord), not RE2's ASCII-only defaults. See struct_tag_re in
// scripts/lib/gomask.py for the reference pattern and its review history.
func isStructTag(content []rune) bool {
	n := len(content)
	i := 0

	skipSpace := func() {
		for i < n && pysem.IsSpace(content[i]) {
			i++
		}
	}

	isTagStartRune := func(r rune) bool {
		return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_'
	}

	// matchPair consumes one `key:"value"` pair starting at the current
	// position. It returns false (without side effects the caller cannot
	// tolerate) if no pair is present.
	matchPair := func() bool {
		if i >= n || !isTagStartRune(content[i]) {
			return false
		}
		i++
		for i < n && pysem.IsWord(content[i]) {
			i++
		}
		if i >= n || content[i] != ':' {
			return false
		}
		i++
		if i >= n || content[i] != '"' {
			return false
		}
		i++
		for i < n && content[i] != '"' {
			i++
		}
		if i >= n || content[i] != '"' {
			return false
		}
		i++
		return true
	}

	skipSpace()
	if !matchPair() {
		return false
	}
	for {
		mark := i
		spaceRun := 0
		for i < n && pysem.IsSpace(content[i]) {
			i++
			spaceRun++
		}
		if spaceRun == 0 {
			break
		}
		if !matchPair() {
			i = mark
			break
		}
	}
	skipSpace()
	return i == n
}
