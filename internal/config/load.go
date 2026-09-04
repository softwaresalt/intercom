package config

import (
	"fmt"
	"sort"
	"strings"

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
// oracle). Decode performs, in order:
//
//  1. cfg := Default() — pre-populated baseline.
//  2. toml.Decode(data, cfg) — syntax/type errors become apperr.KindConfig.
//  5. md.Undecoded() -> Report.UnknownKeys: sorted, sanitized, capped.
//
// Case-fold collision rejection, the required-key check, and Validate
// integration are wired in by later units (B2, C1) per the Decode
// Contract.
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

	return cfg, report, nil
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

// filterLeafKeys drops any key that is a strict ancestor (dotted-path
// prefix) of another key in the same set.
func filterLeafKeys(keys []toml.Key) []toml.Key {
	leaves := make([]toml.Key, 0, len(keys))
	for i, k := range keys {
		isAncestor := false
		for j, other := range keys {
			if i == j {
				continue
			}
			if len(other) > len(k) && keyIsPrefix(k, other) {
				isAncestor = true
				break
			}
		}
		if !isAncestor {
			leaves = append(leaves, k)
		}
	}
	return leaves
}

// keyIsPrefix reports whether prefix is a segment-wise prefix of full.
func keyIsPrefix(prefix, full toml.Key) bool {
	if len(prefix) > len(full) {
		return false
	}
	for i := range prefix {
		if prefix[i] != full[i] {
			return false
		}
	}
	return true
}

// sanitizeKeyPath truncates a dotted TOML key path (as rendered by
// toml.Key.String(), which already renders any character requiring
// quoting — including control characters such as a raw newline — through
// its own Go-syntax-style quoting/escaping) to maxUnknownKeyBytes.
// Any residual raw ASCII control character is additionally escaped here as
// defense in depth, in case a future decoder version renders one
// unescaped.
func sanitizeKeyPath(path string) string {
	var b strings.Builder
	for _, r := range path {
		if r < 0x20 || r == 0x7f {
			fmt.Fprintf(&b, `\x%02x`, r)
			continue
		}
		b.WriteRune(r)
	}

	escaped := b.String()
	if len(escaped) > maxUnknownKeyBytes {
		escaped = escaped[:maxUnknownKeyBytes]
	}
	return escaped
}
