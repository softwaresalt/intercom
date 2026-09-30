// This file (tomlfallback.go) ports strip_toml_comment,
// scan_toml_with_fallback and the shared bare_key_re-based line lexer from
// scripts/lib/retired_arch.py. It is the defensive fallback engine used as
// the real scan path only when a tomllib-equivalent parser is unavailable,
// but --self-test exercises it directly against every fixture regardless,
// so it is never untested dead code.
package retiredarch

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// bareKeyRe ports bare_key_re = re.compile(r'[A-Za-z_][A-Za-z0-9_]*'). No
// \b boundary handling is needed here (unlike goIdentifierRe): Python uses
// plain .findall(), never \b, on this pattern.
var bareKeyRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// tomlLineState ports the `state` dict threaded through strip_toml_comment
// and scan_toml_with_fallback: the simplified per-line lexer's carried
// quote/multiline-string state.
type tomlLineState struct {
	inBasic            bool
	inLiteral          bool
	inMultilineBasic   bool
	inMultilineLiteral bool
}

// runesHavePrefix reports whether runes[i:] starts with prefix (compared
// rune-for-rune; prefix is always ASCII here).
func runesHavePrefix(runes []rune, i int, prefix string) bool {
	p := []rune(prefix)
	if i+len(p) > len(runes) {
		return false
	}
	for k, r := range p {
		if runes[i+k] != r {
			return false
		}
	}
	return true
}

// stripTomlComment ports strip_toml_comment verbatim, operating on runes
// (not bytes) since Python indexes line by codepoint.
func stripTomlComment(line string, state *tomlLineState) string {
	runes := []rune(line)
	n := len(runes)
	var out []rune
	i := 0
	for i < n {
		if state.inMultilineBasic {
			if runesHavePrefix(runes, i, `"""`) {
				backslashes := 0
				j := i - 1
				for j >= 0 && runes[j] == '\\' {
					backslashes++
					j--
				}
				if backslashes%2 == 0 {
					out = append(out, '"', '"', '"')
					i += 3
					state.inMultilineBasic = false
					continue
				}
			}
			out = append(out, runes[i])
			i++
			continue
		}

		if state.inMultilineLiteral {
			if runesHavePrefix(runes, i, "'''") {
				out = append(out, '\'', '\'', '\'')
				i += 3
				state.inMultilineLiteral = false
				continue
			}
			out = append(out, runes[i])
			i++
			continue
		}

		ch := runes[i]

		if state.inBasic {
			out = append(out, ch)
			if ch == '\\' && i+1 < n {
				out = append(out, runes[i+1])
				i += 2
				continue
			}
			if ch == '"' {
				state.inBasic = false
			}
			i++
			continue
		}

		if state.inLiteral {
			out = append(out, ch)
			if ch == '\'' {
				state.inLiteral = false
			}
			i++
			continue
		}

		if runesHavePrefix(runes, i, `"""`) {
			out = append(out, '"', '"', '"')
			i += 3
			state.inMultilineBasic = true
			continue
		}
		if runesHavePrefix(runes, i, "'''") {
			out = append(out, '\'', '\'', '\'')
			i += 3
			state.inMultilineLiteral = true
			continue
		}
		if ch == '#' {
			break
		}
		out = append(out, ch)
		switch ch {
		case '"':
			state.inBasic = true
		case '\'':
			state.inLiteral = true
		}
		i++
	}
	return string(out)
}

// scanTomlFallback ports scan_toml_with_fallback. It returns an error only
// for a file-read/UTF-8-decode failure (ED-2 territory, handled by the
// caller); every TOML-content-level problem it detects itself (an
// unterminated string at EOF) is reported as an ordinary fail-closed
// finding, exactly as Python does.
func scanTomlFallback(path string) ([]string, error) {
	text, err := pysem.ReadText(path)
	if err != nil {
		return nil, err
	}
	posixPath := filepath.ToSlash(path)

	var findings []string
	var currentTable []string
	state := &tomlLineState{}

	for i, rawLine := range pysem.SplitLines(text) {
		lineNo := i + 1

		// 015.009-T (V6, ported verbatim): capture the multiline flags
		// BEFORE stripTomlComment mutates them for THIS line.
		preLineMultiline := state.inMultilineBasic || state.inMultilineLiteral
		line := pysem.Strip(stripTomlComment(rawLine, state))
		if line == "" {
			continue
		}

		if !preLineMultiline && strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			header := pysem.Strip(strings.Trim(line, "[]"))
			currentTable = bareKeyRe.FindAllString(header, -1)
			if token, model, ok := matchesForbiddenParts(composeTomlParts(currentTable)); ok {
				findings = append(findings, fmt.Sprintf(
					"%s:%d: retired token %s in TOML table key %s (via %s model)",
					posixPath, lineNo, pysem.Repr(token), pysem.Repr(strings.Join(currentTable, ".")), model,
				))
			}
			continue
		}

		if preLineMultiline || !strings.Contains(line, "=") {
			continue
		}
		lhs := strings.SplitN(line, "=", 2)[0]
		keys := make([]string, 0, len(currentTable)+2)
		keys = append(keys, currentTable...)
		keys = append(keys, bareKeyRe.FindAllString(lhs, -1)...)
		if token, model, ok := matchesForbiddenParts(composeTomlParts(keys)); ok {
			findings = append(findings, fmt.Sprintf(
				"%s:%d: retired token %s in TOML key %s (via %s model)",
				posixPath, lineNo, pysem.Repr(token), pysem.Repr(strings.Join(keys, ".")), model,
			))
		}
	}

	// Fail closed (AC-6): an unterminated string at EOF means this
	// lexer's simplified state tracking cannot vouch for the rest of the
	// file.
	if state.inMultilineBasic || state.inMultilineLiteral || state.inBasic || state.inLiteral {
		findings = append(findings, fmt.Sprintf("%s: unterminated string at EOF (fail-closed)", posixPath))
	}

	return findings, nil
}
