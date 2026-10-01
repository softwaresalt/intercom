// Package retiredarch reimplements the repository's retired-architecture
// regression gate (scripts/check-retired-architecture.sh /
// the M4-deleted retired_arch module) on top of
// tools/gatecheck/internal/{pysem,gomask}, so the invariant -- no retired
// component name (Slack/socketmode/channel_id/team_id/acp/host_cli/
// ipc_name) survives as a live Go identifier or TOML key under the tracked
// tree -- can be enforced without a Python interpreter.
//
// This file (ident.go) ports split_camel_acronym, split_identifier,
// _VOCAB_WORDS, segment_whole and _segment_whole_exact from
// the M4-deleted retired_arch module, using tools/gatecheck/internal/pysem for
// every Python-str-semantics decision (isdigit/isupper/islower/lower)
// rather than Go's ASCII-biased unicode helpers directly (plan C-4).
package retiredarch

import (
	"sort"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// splitCamelAcronym segments one underscore-delimited chunk into
// camelCase/acronym components. Ported verbatim (including the
// acronym-then-plural-suffix disambiguation and the inert-separator
// fallback for stray non-cased runes such as '.' encountered via a quoted
// TOML key) from split_camel_acronym in the M4-deleted retired_arch module.
//
// Operates on runes (not bytes): Python indexes chunk by codepoint, and a
// byte-indexed walk would misclassify multi-byte UTF-8 continuation bytes.
func splitCamelAcronym(chunk string) []string {
	runes := []rune(chunk)
	n := len(runes)
	var tokens []string
	i := 0
	for i < n {
		ch := runes[i]
		if pysem.IsDigit(ch) {
			j := i
			for j < n && pysem.IsDigit(runes[j]) {
				j++
			}
			tokens = append(tokens, string(runes[i:j]))
			i = j
			continue
		}
		if pysem.IsUpper(ch) {
			j := i
			for j < n && pysem.IsUpper(runes[j]) {
				j++
			}
			runLen := j - i
			if runLen == 1 {
				k := j
				for k < n && pysem.IsLower(runes[k]) {
					k++
				}
				tokens = append(tokens, string(runes[i:k]))
				i = k
				continue
			}
			if j < n && pysem.IsLower(runes[j]) {
				k := j
				for k < n && pysem.IsLower(runes[k]) {
					k++
				}
				tail := string(runes[j:k])
				if tail == "s" || tail == "es" {
					tokens = append(tokens, string(runes[i:k]))
				} else {
					tokens = append(tokens, string(runes[i:j-1]))
					tokens = append(tokens, string(runes[j-1:k]))
				}
				i = k
				continue
			}
			tokens = append(tokens, string(runes[i:j]))
			i = j
			continue
		}
		j := i
		for j < n && pysem.IsLower(runes[j]) {
			j++
		}
		if j == i {
			// Fix (post-015.008-T review, ported verbatim): ch matched
			// none of digit/upper/lower (e.g. '.', ' ', ':', or another
			// non-cased character). Reachable via decomposeTomlKey on a
			// quoted TOML key. Treat it as an inert single-rune
			// separator and always advance i by exactly one.
			i++
			continue
		}
		tokens = append(tokens, string(runes[i:j]))
		i = j
	}
	return tokens
}

// splitIdentifier ports split_identifier: it splits name on '_', segments
// each chunk with splitCamelAcronym, and lowercases every resulting token
// via pysem.Lower (not Go's ASCII/simple-casefold path -- see pysem.Lower's
// doc comment for the İ/Final-Sigma special cases this must reproduce).
// If every chunk yields zero tokens (e.g. name is entirely underscores),
// the fallback is a single token: pysem.Lower(name).
func splitIdentifier(name string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(name); i++ {
		if i == len(name) || name[i] == '_' {
			chunk := name[start:i]
			for _, token := range splitCamelAcronym(chunk) {
				parts = append(parts, pysem.Lower(token))
			}
			start = i + 1
		}
	}
	if len(parts) == 0 {
		return []string{pysem.Lower(name)}
	}
	return parts
}

// vocabWords ports _VOCAB_WORDS: the deduplicated set of every word
// appearing in forbiddenParts' values, sorted by rune length descending.
//
// NON-DETERMINISM NOTE (discovered during M2-T1 golden capture, not a
// migration-introduced divergence): Python's own _VOCAB_WORDS is built as
// `sorted({...set comprehension...}, key=len, reverse=True)`. Ties (words
// of equal length) are broken by the underlying set's iteration order,
// which depends on CPython's per-process string-hash randomization
// (PYTHONHASHSEED) -- confirmed empirically: six fresh, independent
// `python -c "..."` invocations printed six DIFFERENT tie-broken
// orderings of _VOCAB_WORDS. This is a pre-existing property of the reference
// implementation itself, not something this port introduces. It was also
// confirmed NOT to affect any observable finding: scanning every
// scripts/testdata/retiredgo{,-differential}/*.go and
// scripts/testdata/retired-*.toml fixture under tomllib+fallback across
// eight fresh Python processes produced byte-identical combined output
// every time, despite _VOCAB_WORDS' tie order varying across those same
// eight processes. Go's map/slice iteration is not hash-randomized the
// same way Python's is, so this port pins ONE fixed, reproducible
// tie-break (first-occurrence order when flattening forbiddenParts in
// its own declared order, via a stable sort) rather than reproducing
// Python's randomness -- any of the observed Python orderings is
// equally "correct" per the above proof, so this is a documented
// non-issue, not an enumerated delta.
var vocabWords = buildVocabWords()

func buildVocabWords() []string {
	seen := make(map[string]bool, 16)
	words := make([]string, 0, 16)
	for _, fp := range forbiddenParts {
		for _, w := range fp.want {
			if !seen[w] {
				seen[w] = true
				words = append(words, w)
			}
		}
	}
	sort.SliceStable(words, func(i, j int) bool {
		return utf8.RuneCountInString(words[i]) > utf8.RuneCountInString(words[j])
	})
	return words
}

// segmentWholeExact is the exact-match DP core for segmentWhole (no plural
// tolerance), ported verbatim from _segment_whole_exact.
func segmentWholeExact(word string) [][]string {
	runes := []rune(word)
	n := len(runes)
	dp := make([][][]string, n+1)
	dp[0] = [][]string{{}}
	for end := 1; end <= n; end++ {
		var segmentations [][]string
		for _, tok := range vocabWords {
			tokLen := utf8.RuneCountInString(tok)
			start := end - tokLen
			if start >= 0 && len(dp[start]) > 0 && string(runes[start:end]) == tok {
				for _, prefix := range dp[start] {
					seg := make([]string, len(prefix)+1)
					copy(seg, prefix)
					seg[len(prefix)] = tok
					segmentations = append(segmentations, seg)
				}
			}
		}
		dp[end] = segmentations
	}
	return dp[n]
}

// segmentWhole ports segment_whole: every way word can be segmented
// EXACTLY and COMPLETELY into vocabWords, retrying with a trailing 's'/'es'
// suffix stripped (the fused-plural tolerance) when the exact attempt on
// the full word fails. Returns nil (not just an empty, non-nil slice) when
// no segmentation exists, matching Python's `[]` (falsy, and this port
// never distinguishes nil from an empty allocated slice for callers, which
// only ever range over the result).
func segmentWhole(word string) [][]string {
	if exact := segmentWholeExact(word); len(exact) > 0 {
		return exact
	}
	for _, suffix := range []string{"es", "s"} {
		if len(word) >= len(suffix) && word[len(word)-len(suffix):] == suffix &&
			utf8.RuneCountInString(word) > utf8.RuneCountInString(suffix) {
			stripped := word[:len(word)-len(suffix)]
			if s := segmentWholeExact(stripped); len(s) > 0 {
				return s
			}
		}
	}
	return nil
}
