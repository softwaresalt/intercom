package config

import (
	"fmt"
	"os"
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
//  3. Case-fold collision rejection — two distinct key paths equal under
//     strings.EqualFold (divergence V2).
//  4. Required-key check — md.IsDefined("default_workspace_root") only;
//     host_cli's required-ness is enforced solely by validation rule 6
//     (Decision R8), which covers both the absent and explicitly-empty
//     case with one message.
//  5. md.Undecoded() -> Report.UnknownKeys: sorted, sanitized, capped.
//  6. cfg.Validate() -> its Report merged into the returned Report.
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
func findCaseFoldCollision(keys []toml.Key) ([2]string, bool) {
	seen := make(map[string]string, len(keys))
	for _, k := range keys {
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

// Load stats path, rejects anything larger than MaxConfigBytes, reads it,
// and delegates to Decode.
func Load(path string) (*Config, Report, error) {
	var report Report

	info, err := os.Stat(path)
	if err != nil {
		return nil, report, apperr.Newf(apperr.KindConfig, "cannot read config file '%s': %s — create it or pass --config with a valid path", path, err.Error())
	}
	if info.Size() > MaxConfigBytes {
		return nil, report, apperr.Newf(apperr.KindConfig, "config file '%s' exceeds the maximum allowed size of %d bytes", path, int64(MaxConfigBytes))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, report, apperr.Newf(apperr.KindConfig, "cannot read config file '%s': %s — create it or pass --config with a valid path", path, err.Error())
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
