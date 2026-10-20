package config

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// remoteKnownKeys are the schema keys under [remote]. They are exempt from
// credential-shape screening (token_lifetime_seconds legitimately contains
// "token" but is a plain duration).
var remoteKnownKeys = []string{
	"token_lifetime_seconds",
	"allowed_origins",
	"allowed_hosts",
	"max_connections",
	"remote_approve_destructive",
}

// credentialFragments are substrings (matched against the lowercased key
// with non-alphanumerics removed) that mark a key as credential-shaped.
// The screen is deliberately broad: it fails closed, because credentials
// must never come from config.toml.
var credentialFragments = []string{
	"token", "secret", "passw", "pwd", "credential", "bearer",
	"key", "auth", "jwt", "cookie", "signature", "session",
}

// isCredentialShapedKey reports whether a single key segment looks like it
// names a credential.
func isCredentialShapedKey(segment string) bool {
	var b strings.Builder
	for _, r := range strings.ToLower(segment) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	norm := b.String()
	for _, f := range credentialFragments {
		if strings.Contains(norm, f) {
			return true
		}
	}
	return false
}

func isKnownRemoteKey(segment string) bool {
	for _, k := range remoteKnownKeys {
		if strings.EqualFold(segment, k) {
			return true
		}
	}
	return false
}

// rejectRemoteCredentialKeys scans every key path present in the document
// and fails closed on any credential-shaped key at or below [remote]. The
// error names the key path only; values are never read or echoed.
func rejectRemoteCredentialKeys(keys []toml.Key) error {
	for _, k := range keys {
		if len(k) < 2 || !strings.EqualFold(k[0], "remote") {
			continue
		}
		if len(k) == 2 && isKnownRemoteKey(k[1]) {
			continue
		}
		for _, seg := range k[1:] {
			if isCredentialShapedKey(seg) {
				return apperr.Newf(apperr.KindConfig,
					"%s is a credential-shaped key; credentials must never be stored in config.toml (remove it from [remote])",
					sanitizeKeyPath(k.String()))
			}
		}
	}
	return nil
}

// validateRemote implements the [remote] rule group, executed after the
// seven shipped rules. Messages are field-specific and never echo entry
// values (an origin may carry userinfo).
func validateRemote(r RemoteConfig) error {
	if r.TokenLifetimeSeconds < MinRemoteTokenLifetimeSeconds || r.TokenLifetimeSeconds > MaxRemoteTokenLifetimeSeconds {
		return apperr.Newf(apperr.KindConfig, "remote.token_lifetime_seconds must be between %d and %d",
			MinRemoteTokenLifetimeSeconds, MaxRemoteTokenLifetimeSeconds)
	}
	seen := make(map[string]struct{}, len(r.AllowedOrigins))
	for i, o := range r.AllowedOrigins {
		if reason := originProblem(o); reason != "" {
			return apperr.Newf(apperr.KindConfig, "remote.allowed_origins[%d] %s", i, reason)
		}
		if _, dup := seen[o]; dup {
			return apperr.Newf(apperr.KindConfig, "remote.allowed_origins[%d] is a duplicate entry", i)
		}
		seen[o] = struct{}{}
	}
	seen = make(map[string]struct{}, len(r.AllowedHosts))
	for i, h := range r.AllowedHosts {
		if reason := hostProblem(h); reason != "" {
			return apperr.Newf(apperr.KindConfig, "remote.allowed_hosts[%d] %s", i, reason)
		}
		if _, dup := seen[h]; dup {
			return apperr.Newf(apperr.KindConfig, "remote.allowed_hosts[%d] is a duplicate entry", i)
		}
		seen[h] = struct{}{}
	}
	if r.MaxConnections < 1 {
		return apperr.New(apperr.KindConfig, "remote.max_connections must be at least 1")
	}
	return nil
}

const originForm = "must be an exact lowercase http(s)://host[:port] origin"

// originProblem returns a non-empty reason if o is not an exact origin. The
// reason never includes o itself (it may carry userinfo).
func originProblem(o string) string {
	if o == "" {
		return "must not be empty"
	}
	if strings.Contains(o, "*") {
		return "must not contain a wildcard"
	}
	if o != strings.ToLower(o) {
		return originForm + " (uppercase is not allowed)"
	}
	for _, r := range o {
		if r <= ' ' || r == 0x7f {
			return originForm + " (whitespace/control characters are not allowed)"
		}
	}
	u, err := url.Parse(o)
	if err != nil {
		return originForm
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return originForm + " (scheme must be http or https)"
	}
	if u.User != nil || strings.Contains(u.Host, "@") {
		return originForm + " (userinfo is not allowed)"
	}
	if u.Path != "" || u.RawPath != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || strings.ContainsAny(o, "?#") {
		return originForm + " (no path, query, or fragment)"
	}
	if u.Hostname() == "" {
		return originForm + " (host is required)"
	}
	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return originForm + " (port must be 1-65535)"
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return originForm + " (empty port)"
	}
	if hostProblem(u.Hostname()) != "" && net.ParseIP(u.Hostname()) == nil {
		return originForm + " (invalid host)"
	}
	return ""
}

// hostProblem returns a non-empty reason if h is not a plain lowercase
// host name (or IPv4 literal) without scheme, port, path, or wildcard.
func hostProblem(h string) string {
	if h == "" {
		return "must not be empty"
	}
	if strings.Contains(h, "*") {
		return "must not contain a wildcard"
	}
	if h != strings.ToLower(h) {
		return "must be a lowercase host name (uppercase is not allowed)"
	}
	for _, r := range h {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-'
		if !ok {
			return "must be a plain host name (no scheme, port, path, userinfo, or whitespace)"
		}
	}
	if strings.HasPrefix(h, ".") || strings.HasSuffix(h, ".") || strings.HasPrefix(h, "-") || strings.Contains(h, "..") {
		return "must be a well-formed host name"
	}
	return ""
}
