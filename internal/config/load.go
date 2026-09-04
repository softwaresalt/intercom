package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/BurntSushi/toml"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// maxUnknownKeyBytes bounds the length of a single sanitized unknown-key
// path reported in Report.UnknownKeys.
const maxUnknownKeyBytes = 128

// maxUnknownKeys bounds the number of unknown-key entries reported; beyond
// this a single "… N more" marker is appended instead of every remaining
// entry (security hardening against a hostile config forging many bogus
// keys — B1 acceptance iv).
const maxUnknownKeys = 64

// Decode parses TOML data into a *Config pre-populated by Default(),
// tolerant of unknown keys (requirement R3: no deny_unknown_fields in the
// oracle). Decode performs these checks, conceptually in this order:
//
//  1. cfg := Default() — pre-populated baseline.
//  2. toml.Decode(data, cfg) — syntax/type errors become apperr.KindConfig.
//  3. Case-fold collision rejection — two distinct key paths equal under
//     strings.EqualFold (divergence V2).
//  4. Required-key check — md.IsDefined("default_workspace_root") only;
//     host_cli's required-ness is enforced solely by validation rule 6
//     (Decision R8), which covers both the absent and explicitly-empty
//     case with one message.
//  5. md.Undecoded() -> Report.UnknownKeys: sorted, sanitized, capped.
//  6. cfg.Validate() -> its Report merged into the returned Report.
//
// Step 5 is actually computed in the implementation immediately after
// step 2, ahead of steps 3-4, even though it is conceptually presented
// last: this has no observable side effects on steps 3-4's own outcome,
// and computing it early ensures Report.UnknownKeys is still populated on
// the report-on-error path below when step 3 or 4 fails, rather than
// being silently dropped by an early return.
//
// Order-of-checks caveat: step 4's required-key check for
// default_workspace_root's ABSENCE runs, and can return, before step 6
// (cfg.Validate(), whose own 10 rules run in the fixed order documented
// in docs/config-reference.md) is ever invoked. A document that BOTH
// omits default_workspace_root AND violates a rule ordered before
// Validate's rule 2 (e.g. rule 1, max_concurrent_sessions == 0) therefore
// still reports step 4's "default_workspace_root must be set" from
// Decode, not Validate's rule-1 message — Validate's own fixed-order
// guarantee applies fully only when Validate is called directly, not
// through this earlier Decode-level gate. This is deliberate (Decision
// R8's design predates and governs this behavior), locked by
// TestDecodeRequiredKeyPrecheckPrecedesValidate in decode_test.go.
//
// Report-on-error semantics: Decode returns the Report populated with
// everything gathered so far even when it also returns a non-nil error, so
// a caller can log unknown keys alongside a failure. A returned *Config is
// only valid when err == nil.
func Decode(data string) (*Config, Report, error) {
	cfg := Default()
	var report Report

	md, err := toml.Decode(data, cfg)
	if err != nil {
		return cfg, report, apperr.Wrap(apperr.KindConfig, err)
	}

	report.UnknownKeys = sanitizeUnknownKeys(md.Undecoded())

	if collision, ok := findCaseFoldCollision(md.Keys()); ok {
		return cfg, report, apperr.Newf(apperr.KindConfig, "config keys %q and %q differ only by case; rename one to avoid ambiguous decoding", collision[0], collision[1])
	}

	if !md.IsDefined("default_workspace_root") {
		return cfg, report, apperr.New(apperr.KindConfig, "default_workspace_root must be set")
	}

	validateReport, err := cfg.Validate()
	report.Warnings = append(report.Warnings, validateReport.Warnings...)
	if err != nil {
		return cfg, report, err
	}

	return cfg, report, nil
}

// findCaseFoldCollision scans every distinct key path present in the
// document (at every nesting level) and reports the first pair equal
// under strings.EqualFold but not identical — e.g. "host_cli" and
// "Host_CLI" both target the same struct field, and because the decoder
// iterates a Go map internally, the winner would otherwise be
// nondeterministic (divergence V2, security finding P1-f). A lone
// non-canonical spelling with no colliding variant is unaffected
// (divergence V2b).
//
// Key paths that fall inside a map-typed schema field (Commands,
// Slack.MarkdownUploadExtensions) are excluded: their sub-keys are
// literal, case-sensitive runtime data assigned directly into a Go map,
// never resolved through the decoder's ambiguous case-fold struct-field
// fallback, so two differently-cased map entries (e.g. two distinct
// named commands) are not a collision (review finding: over-rejection of
// valid tolerant-decode input, requirement R3).
//
// Known accepted limitation: toml.MetaData.Keys() reports an
// array-of-tables path (e.g. "workspace.workspace_id") identically for
// every [[workspace]] entry, with no per-instance index. Two different
// entries using different case for the same field (e.g. entry 1's
// workspace_id vs entry 2's Workspace_ID) are therefore indistinguishable
// from a true same-instance collision using only this API and are still
// rejected. This is a narrow, accepted divergence from ideal behavior —
// no known library-level mechanism disambiguates per-instance key
// identity — rather than a defect to be fixed here.
func findCaseFoldCollision(keys []toml.Key) ([2]string, bool) {
	mapPaths := mapFieldPaths()

	seen := make(map[string]string, len(keys))
	for _, k := range keys {
		if underMapField(k, mapPaths) {
			continue
		}

		path := k.String()
		folded := strings.ToLower(path)
		if existing, ok := seen[folded]; ok {
			if existing != path {
				return [2]string{existing, path}, true
			}
			continue
		}
		seen[folded] = path
	}
	return [2]string{}, false
}

// mapFieldPaths returns the dotted schema paths (as segment slices) of
// every map-typed field in the Config struct hierarchy.
func mapFieldPaths() [][]string {
	return collectMapPaths(reflect.TypeOf(Config{}), nil)
}

// collectMapPaths walks t's toml-tagged fields, recursing into nested
// struct fields and the element type of a slice of structs, and returns
// the path of every field whose Go type is a map.
func collectMapPaths(t reflect.Type, prefix []string) [][]string {
	var paths [][]string
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("toml")
		if tag == "" || tag == "-" {
			continue
		}

		path := make([]string, len(prefix), len(prefix)+1)
		copy(path, prefix)
		path = append(path, tag)

		switch field.Type.Kind() {
		case reflect.Map:
			paths = append(paths, path)
		case reflect.Struct:
			paths = append(paths, collectMapPaths(field.Type, path)...)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.Struct {
				paths = append(paths, collectMapPaths(field.Type.Elem(), path)...)
			}
		}
	}
	return paths
}

// underMapField reports whether k's segments extend beyond one of
// mapPaths — i.e. k names an entry stored inside a map-typed field rather
// than the map field itself.
func underMapField(k toml.Key, mapPaths [][]string) bool {
	for _, mp := range mapPaths {
		if len(k) > len(mp) && segmentsEqualFold(k[:len(mp)], mp) {
			return true
		}
	}
	return false
}

// segmentsEqualFold reports whether a and b have the same length and are
// equal element-wise under strings.EqualFold.
func segmentsEqualFold(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !strings.EqualFold(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Load stats path, rejects anything larger than MaxConfigBytes, reads it,
// and delegates to Decode.
//
// The size guard reads through an io.LimitReader capped at
// MaxConfigBytes+1 rather than trusting a preceding os.Stat, so a file
// grown by a concurrent writer between the existence check and the read
// cannot smuggle more than MaxConfigBytes bytes past the guard (divergence
// V9 is enforced against what is actually read, not a stale stat result).
func Load(path string) (*Config, Report, error) {
	var report Report

	f, err := os.Open(path)
	if err != nil {
		return nil, report, apperr.Newf(apperr.KindConfig, "cannot read config file '%s': %s — create it or pass --config with a valid path", path, err.Error())
	}
	defer func() {
		_ = f.Close()
	}()

	data, err := io.ReadAll(io.LimitReader(f, MaxConfigBytes+1))
	if err != nil {
		return nil, report, apperr.Newf(apperr.KindConfig, "cannot read config file '%s': %s — create it or pass --config with a valid path", path, err.Error())
	}
	if len(data) > MaxConfigBytes {
		return nil, report, apperr.Newf(apperr.KindConfig, "config file '%s' exceeds the maximum allowed size of %d bytes", path, int64(MaxConfigBytes))
	}

	return Decode(string(data))
}

// sanitizeUnknownKeys sorts, sanitizes, and caps the undecoded key list per
// the Report.UnknownKeys contract. Undecoded() reports both an entire
// unmatched table and every one of its leaves (e.g. both "nonsense" and
// "nonsense.x" for an unmatched [nonsense] section); filterLeafKeys keeps
// only the most specific (leaf) entries so a caller does not see redundant
// ancestor paths.
func sanitizeUnknownKeys(keys []toml.Key) []string {
	leaves := filterLeafKeys(keys)

	paths := make([]string, 0, len(leaves))
	for _, k := range leaves {
		paths = append(paths, sanitizeKeyPath(k.String()))
	}
	sort.Strings(paths)

	if len(paths) > maxUnknownKeys {
		more := len(paths) - maxUnknownKeys
		paths = paths[:maxUnknownKeys]
		paths = append(paths, fmt.Sprintf("… %d more", more))
	}

	return paths
}

// filterLeafKeys drops any key that is a strict ancestor (segment-wise
// prefix) of another key in the same set, in O(n log n): each key's
// segments are joined with a control-character separator guaranteed not
// to appear in a TOML key (NUL), the joined forms are sorted, and a
// single forward pass detects ancestors by comparing each entry only to
// its immediate successor. This relies on the standard prefix property of
// lexicographic ordering — every proper extension of a string sorts
// immediately after it with no unrelated entry interleaved — so a single
// adjacent comparison is sufficient without the O(n^2) all-pairs scan a
// hostile config with many thousands of unknown keys would otherwise
// force (security finding: unbounded CPU cost ahead of the maxUnknownKeys
// output cap).
func filterLeafKeys(keys []toml.Key) []toml.Key {
	const separator = "\x00"

	type joined struct {
		key  toml.Key
		path string
	}

	entries := make([]joined, len(keys))
	for i, k := range keys {
		entries[i] = joined{key: k, path: strings.Join(k, separator)}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })

	leaves := make([]toml.Key, 0, len(entries))
	for i, cur := range entries {
		isAncestor := i+1 < len(entries) &&
			len(entries[i+1].path) > len(cur.path) &&
			entries[i+1].path[:len(cur.path)] == cur.path &&
			entries[i+1].path[len(cur.path):len(cur.path)+len(separator)] == separator
		if !isAncestor {
			leaves = append(leaves, cur.key)
		}
	}
	return leaves
}

// sanitizeKeyPath escapes any residual raw ASCII control character in a
// dotted TOML key path (defense in depth — toml.Key.String() already
// renders any character requiring quoting, including control characters
// such as a raw newline, through its own Go-syntax-style
// quoting/escaping, so this should rarely trigger against the current
// decoder version) and truncates the result to maxUnknownKeyBytes on a
// valid rune boundary, so a multi-byte UTF-8 character straddling the
// byte limit is never split into an invalid partial encoding (a raw
// byte-index slice could otherwise corrupt the string for any caller that
// logs or JSON-marshals it).
func sanitizeKeyPath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if r < 0x20 || r == 0x7f {
			fmt.Fprintf(&b, `\x%02x`, r)
			continue
		}
		b.WriteRune(r)
	}

	return truncateAtRuneBoundary(b.String(), maxUnknownKeyBytes)
}

// truncateAtRuneBoundary returns the longest prefix of s no longer than
// maxBytes that ends on a complete UTF-8 rune boundary.
func truncateAtRuneBoundary(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}
