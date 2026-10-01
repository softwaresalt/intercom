// This file (match.go) ports forbidden_parts, window_matches,
// matches_forbidden_sequence, matches_forbidden_concat and
// matches_forbidden_parts from the M4-deleted retired_arch module.
package retiredarch

import "strings"

// forbiddenPart is one (token, want-sequence) entry from Python's
// forbidden_parts dict.
type forbiddenPart struct {
	token string
	want  []string
}

// forbiddenParts is an ORDERED slice, not a map: matchesForbiddenSequence
// and matchesForbiddenConcat must iterate it in exactly this declaration
// order to reproduce Python's dict-insertion-order "first match wins"
// semantics (forbidden_parts.items() in a CPython 3.7+ dict iterates in
// insertion order, which is NOT hash-randomized the way a set's order is).
var forbiddenParts = []forbiddenPart{
	{"slack", []string{"slack"}},
	{"socketmode", []string{"socket", "mode"}},
	{"channel_id", []string{"channel", "id"}},
	{"team_id", []string{"team", "id"}},
	{"acp", []string{"acp"}},
	{"host_cli", []string{"host", "cli"}},
	{"ipc_name", []string{"ipc", "name"}},
}

// windowMatches ports window_matches: a forbidden sequence `want` matches
// `parts` if some contiguous window of `parts` equals `want`, except that
// the final element of the window may carry a plural 's'/'es' suffix
// relative to the final element of `want`.
func windowMatches(parts, want []string) bool {
	span := len(want)
	for i := 0; i+span <= len(parts); i++ {
		window := parts[i : i+span]
		eq := true
		for k := 0; k < span-1; k++ {
			if window[k] != want[k] {
				eq = false
				break
			}
		}
		if !eq {
			continue
		}
		last, wantLast := window[span-1], want[span-1]
		if last == wantLast {
			return true
		}
		for _, suffix := range []string{"es", "s"} {
			if strings.HasSuffix(last, suffix) && strings.TrimSuffix(last, suffix) == wantLast {
				return true
			}
		}
	}
	return false
}

// matchesForbiddenSequence ports matches_forbidden_sequence: the
// camel/acronym-boundary model, matched against the case-boundary-derived
// component list produced by splitIdentifier.
func matchesForbiddenSequence(parts []string) (token string, ok bool) {
	for _, fp := range forbiddenParts {
		if len(parts) >= len(fp.want) && windowMatches(parts, fp.want) {
			return fp.token, true
		}
	}
	return "", false
}

// matchesForbiddenConcat ports matches_forbidden_concat: the no-separator
// concat model. For each individual component splitIdentifier could not
// further decompose by case boundary, attempt a full whole-part
// decomposition (segmentWhole) and re-apply the window/suffix rule to it.
func matchesForbiddenConcat(parts []string) (token string, ok bool) {
	for _, component := range parts {
		for _, segmentation := range segmentWhole(component) {
			for _, fp := range forbiddenParts {
				if len(segmentation) >= len(fp.want) && windowMatches(segmentation, fp.want) {
					return fp.token, true
				}
			}
		}
	}
	return "", false
}

// matchesForbiddenParts ports matches_forbidden_parts: the single dispatch
// point for the two matching models, returning (token, model, true) so
// findings can report which model fired, or ("", "", false).
func matchesForbiddenParts(parts []string) (token, model string, ok bool) {
	if t, k := matchesForbiddenSequence(parts); k {
		return t, "sequence", true
	}
	if t, k := matchesForbiddenConcat(parts); k {
		return t, "concat", true
	}
	return "", "", false
}
