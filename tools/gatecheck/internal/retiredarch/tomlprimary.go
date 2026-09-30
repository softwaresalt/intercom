// This file (tomlprimary.go) ports decompose_toml_key, compose_toml_parts,
// report_toml_key, walk_toml_value and scan_toml_with_tomllib from
// scripts/lib/retired_arch.py, using github.com/BurntSushi/toml as the
// primary TOML engine (the Go equivalent of CPython's stdlib tomllib).
//
// THE ORDERING PROBLEM (plan C-6): Python's tomllib.loads() returns an
// ordinary dict, which preserves first-insertion order for iteration
// (dict.items()) exactly as the source document declared its keys.
// Go's decoded map[string]interface{} has NO such guarantee (Go map
// iteration order is randomized per-process). BurntSushi's
// toml.MetaData.Keys() DOES return every key path in the document's
// declaration order, but it is a flat list, not a tree, and its exact
// entry-per-key/entry-per-array-element shape must be walked in lockstep
// with the decoded tree to recover sibling and array-element order
// without ever sorting (sorting would silently diverge from Python's
// insertion order and is explicitly prohibited by the plan).
//
// The lockstep cursor algorithm implemented below (tomlCursor plus
// walkTable/walkArrayOfTables/walkArray) was derived empirically against
// BurntSushi/toml v1.6.0 by dumping MetaData.Keys() for a battery of
// shapes (reopened tables via dotted-header continuation, arrays of
// tables, nested arrays of tables, inline tables inside plain arrays,
// mixed scalar/table arrays) and is documented inline at each function.
package retiredarch

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// decomposeTomlKey ports decompose_toml_key: it decomposes a single raw
// TOML key/table-header segment string into splitIdentifier-style
// component tokens, normalising hyphens to underscores at this TOML
// boundary first (bare TOML keys may use hyphens; Go identifiers never
// do).
func decomposeTomlKey(segment string) []string {
	return splitIdentifier(strings.ReplaceAll(segment, "-", "_"))
}

// composeTomlParts ports compose_toml_parts: it flattens decomposeTomlKey
// over every segment of a composed TOML key path (table-prefix segments
// plus the leaf key) into one parts list, so matchesForbiddenParts can see
// a forbidden sequence that only appears once the path is composed.
func composeTomlParts(pathSegments []string) []string {
	var parts []string
	for _, segment := range pathSegments {
		parts = append(parts, decomposeTomlKey(segment)...)
	}
	return parts
}

// reportTomlKey ports report_toml_key's finding-text format. posixPath is
// the already-forward-slashed, root-anchored path (path.as_posix() in
// Python, where path = root / rel_path and root is an absolute path).
func reportTomlKey(posixPath string, keyPath []string, segment, token, model string) string {
	return fmt.Sprintf(
		"%s: retired token %s in TOML key path %s (segment %s) (via %s model)",
		posixPath, pysem.Repr(token), pysem.Repr(strings.Join(keyPath, ".")), pysem.Repr(segment), model,
	)
}

// tomlCursor walks a flat, ordered []toml.Key list (MetaData.Keys()) in
// lockstep with a recursive descent over the decoded tree.
type tomlCursor struct {
	keys []toml.Key
	pos  int
}

// peekChild reports whether the next unconsumed cursor entry extends
// prefix (its first len(prefix) segments equal prefix, and it has at
// least one more segment beyond that). It returns the immediate child key
// name (the segment at index len(prefix)) and whether this entry is an
// EXACT self-entry for that child (length == len(prefix)+1) or a deeper
// entry belonging to a descendant reached via a dotted-header
// continuation that never announced the intermediate table on its own
// (e.g. "[a.b]" implicitly creates "a" with no dedicated Keys() entry for
// "a" itself -- only "a.b" and deeper get entries). A non-exact result is
// NOT consumed here; the caller recurses one level deeper with the same
// cursor position, and each recursion strips one more matched prefix
// segment until the entry finally becomes exact at its own depth.
func (c *tomlCursor) peekChild(prefix []string) (childKey string, exact bool, ok bool) {
	if c.pos >= len(c.keys) {
		return "", false, false
	}
	k := c.keys[c.pos]
	if len(k) <= len(prefix) {
		return "", false, false
	}
	for i, p := range prefix {
		if k[i] != p {
			return "", false, false
		}
	}
	return k[len(prefix)], len(k) == len(prefix)+1, true
}

// peekSelf reports whether the next unconsumed cursor entry is an EXACT
// repeat of prefix itself (length == len(prefix), every segment equal).
// This is the "reopened element" self-entry every array-of-tables element
// after the first requires (see walkArrayOfTables).
func (c *tomlCursor) peekSelf(prefix []string) bool {
	if c.pos >= len(c.keys) {
		return false
	}
	k := c.keys[c.pos]
	if len(k) != len(prefix) {
		return false
	}
	for i, p := range prefix {
		if k[i] != p {
			return false
		}
	}
	return true
}

func (c *tomlCursor) consume() { c.pos++ }

// walkTable walks one decoded TOML table's own fields (a map[string]any),
// driven entirely by the cursor rather than Go's unordered map iteration,
// so sibling order matches the document's declaration order exactly.
//
// The `opened` set (scoped to this single call/table instance, not global)
// guards against double-visiting and double-reporting a child key that
// this table's OWN loop first discovers non-exactly (via a deeper
// dotted-header entry) and later re-encounters as an EXACT "reopened
// table" self-entry (e.g. "[a.b]" ... "[a]" continuing to add siblings of
// "b" under "a": "a" is first discovered non-exactly while opening "b",
// then later the explicit "[a]" header emits its own exact self-entry for
// "a" again -- that second occurrence must resume walking "a"'s fields,
// not re-run the forbidden-match check or re-walk "b").
// A bare array literal's elements (walkArray) never get their own
// boundary marker in Keys(): every element of `arr = [ {...}, {...} ]`
// contributes entries under the exact SAME prefix ("arr"), so once one
// element's fields are exhausted the very next cursor entry (belonging to
// the NEXT element) still "extends prefix" and would otherwise be
// mistaken for one more field of the CURRENT element -- especially when
// two elements share a key name (Keys() then repeats that exact entry
// verbatim, as seen for "arr = [ {x,y}, {y,z} ]"). Bounding the loop by
// `len(opened) == len(table)` (this decoded table's own known field
// count) stops walkTable exactly when its own fields are exhausted,
// before it can wrongly consume a sibling element's entries.
func walkTable(posixPath string, prefix []string, table map[string]interface{}, cursor *tomlCursor, findings *[]string) error {
	opened := make(map[string]bool, len(table))
	for len(opened) < len(table) {
		childKey, exact, ok := cursor.peekChild(prefix)
		if !ok {
			// peekChild rejects an EXACT self-entry for this table's own
			// prefix (length == len(prefix)), which arises when a plain
			// (non-array) table is REOPENED later in the document via a
			// second "[header]" line to add more siblings (e.g. "[a.b]"
			// ... "[a]" continuing with more fields under "a"). That is
			// NOT end-of-table: len(opened) < len(table) still holds,
			// meaning fields declared under the reopened header remain
			// unvisited. Consume that self-entry (mirroring
			// walkArrayOfTables' own analogous reopened-element handling
			// via peekSelf) and keep looping so those fields are still
			// discovered and checked against matchesForbiddenParts,
			// rather than silently returning a false-clean result.
			if cursor.peekSelf(prefix) {
				cursor.consume()
				continue
			}
			// Genuine invariant violation: the cursor has nothing left
			// that extends or repeats this table's own prefix, yet this
			// table still has unvisited fields. Fail closed rather than
			// silently under-reporting (matches this file's existing
			// cursor-desync error style elsewhere).
			return fmt.Errorf("retiredarch: TOML cursor desync: table at %v has %d unvisited field(s) but cursor has no matching entry left", prefix, len(table)-len(opened))
		}
		firstVisit := !opened[childKey]
		if exact {
			cursor.consume()
		}
		child, present := table[childKey]
		if !present {
			return fmt.Errorf("retiredarch: TOML cursor desync: key %q not found in decoded table at %v", childKey, prefix)
		}
		keyPath := make([]string, len(prefix)+1)
		copy(keyPath, prefix)
		keyPath[len(prefix)] = childKey
		if firstVisit {
			opened[childKey] = true
			if token, model, matched := matchesForbiddenParts(composeTomlParts(keyPath)); matched {
				*findings = append(*findings, reportTomlKey(posixPath, keyPath, childKey, token, model))
			}
		}
		if err := walkValue(posixPath, keyPath, child, cursor, findings); err != nil {
			return err
		}
	}
	return nil
}

// walkArrayOfTables walks a TOML array-of-tables value (`[[header]]`
// syntax; decoded by BurntSushi as []map[string]interface{} -- this
// concrete container type never arises any other way, since ordinary
// array-literal syntax always decodes to []interface{}, even when every
// element happens to be an inline table). Element 0's own self-entry was
// already consumed by the parent's walkTable/peekChild call that
// discovered this array; elements 1..N-1 each require consuming one MORE
// repeated self-entry (peekSelf) before their own fields can be walked.
// Nested arrays-of-tables restart this same element-cursor cycle
// independently per parent element (each element gets its own walkTable
// call and thus its own descent into any nested [[prefix.child]] array).
func walkArrayOfTables(posixPath string, prefix []string, arr []map[string]interface{}, cursor *tomlCursor, findings *[]string) error {
	for i, elem := range arr {
		if i > 0 {
			if !cursor.peekSelf(prefix) {
				return fmt.Errorf("retiredarch: TOML cursor desync: expected repeated array-of-tables self-entry at %v (element %d)", prefix, i)
			}
			cursor.consume()
		}
		if err := walkTable(posixPath, prefix, elem, cursor, findings); err != nil {
			return err
		}
	}
	return nil
}

// walkArray walks an ordinary TOML array-literal value (decoded as
// []interface{}), which unlike an array-of-tables never announces its own
// elements in Keys() at all: every element is walked using the SAME
// prefix as the array itself. An inline-table element (map[string]any)
// has its own fields walked via walkTable at that same prefix; a nested
// plain-array element recurses with the same prefix; a scalar element
// consumes nothing. []map[string]interface{} never appears nested inside
// a plain array (TOML's [[header]] array-of-tables syntax only exists at
// the document/table-header level, never as an array-literal value), so
// no case for it is needed here.
func walkArray(posixPath string, prefix []string, arr []interface{}, cursor *tomlCursor, findings *[]string) error {
	for _, elem := range arr {
		switch v := elem.(type) {
		case map[string]interface{}:
			if err := walkTable(posixPath, prefix, v, cursor, findings); err != nil {
				return err
			}
		case []interface{}:
			if err := walkArray(posixPath, prefix, v, cursor, findings); err != nil {
				return err
			}
		default:
			// scalar leaf: no entries, no recursion.
		}
	}
	return nil
}

// walkValue ports the type-dispatch half of walk_toml_value: it routes a
// decoded value to the table/array-of-tables/array walker matching its
// concrete Go type, or does nothing for a scalar leaf.
func walkValue(posixPath string, prefix []string, value interface{}, cursor *tomlCursor, findings *[]string) error {
	switch v := value.(type) {
	case map[string]interface{}:
		return walkTable(posixPath, prefix, v, cursor, findings)
	case []map[string]interface{}:
		return walkArrayOfTables(posixPath, prefix, v, cursor, findings)
	case []interface{}:
		return walkArray(posixPath, prefix, v, cursor, findings)
	default:
		return nil
	}
}

// scanTomlPrimary ports scan_toml_with_tomllib. Python's implementation
// wraps BOTH the file read (path.read_text) and the parse
// (tomllib.loads) in one `except Exception` and formats any failure
// (missing file, invalid UTF-8, or a genuine TOML syntax error) the same
// way -- a single fail-closed finding, never a propagated error -- so this
// port does the same rather than surfacing a Go error for a read failure.
//
// BOM parity (C-6/ED-6 citation): BurntSushi/toml v1.6.0's parse.go
// (lines 43-50) explicitly strips a leading UTF-8 BOM (\xef\xbb\xbf)
// before lexing, i.e. it silently ACCEPTS a BOM-prefixed document.
// Python's tomllib.loads() receives the literal BOM rune (Path.read_text
// does not strip it, unlike the 'utf-8-sig' codec) and fails closed with
// a parse error at line 1 column 1. This port rejects a leading U+FEFF
// explicitly, before ever calling toml.Decode, to reproduce Python's
// fail-closed behavior rather than BurntSushi's silent-accept.
func scanTomlPrimary(path string) []string {
	posixPath := filepath.ToSlash(path)
	text, err := pysem.ReadText(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: TOML parse error (fail-closed): %v", posixPath, err)}
	}
	if strings.HasPrefix(text, "\ufeff") {
		return []string{fmt.Sprintf(
			"%s: TOML parse error (fail-closed): %s",
			posixPath, "leading UTF-8 BOM is not valid TOML under this engine (fail-closed; BurntSushi/toml silently strips it, Python's tomllib does not)",
		)}
	}

	var data map[string]interface{}
	meta, err := toml.Decode(text, &data)
	if err != nil {
		return []string{fmt.Sprintf("%s: TOML parse error (fail-closed): %v", posixPath, err)}
	}

	var findings []string
	cursor := &tomlCursor{keys: meta.Keys()}
	if err := walkTable(posixPath, nil, data, cursor, &findings); err != nil {
		// A cursor-desync error is an internal invariant violation of
		// this port's ordering algorithm, not a TOML content problem;
		// still fail closed rather than silently under-reporting.
		return []string{fmt.Sprintf("%s: TOML parse error (fail-closed): %v", posixPath, err)}
	}
	return findings
}
