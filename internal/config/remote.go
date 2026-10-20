package config

import (
	"net"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// remoteKnownKeys are the schema keys under [remote]. They are exempt from
// credential-shape screening (token_lifetime_seconds legitimately contains
// "token" but is a plain duration, max_connections contains "connection").
//
// NEW [remote] KEYS MUST BE ADDED HERE: the screen below is intentionally a
// superset of ADR MVP-2 section 4's list and fails closed on anything
// credential-shaped, so a new legitimate key whose name trips a fragment is
// rejected until it is listed.
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
	"token", "secret", "pwd", "credential", "bearer",
	"key", "auth", "jwt", "cookie", "signature", "session",
	"pass", "pat", "psk", "otp", "pin", "private", "cert", "hmac",
	"salt", "dsn", "connstr", "connection",
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
	if r.MaxConnections > MaxRemoteConnections {
		return apperr.Newf(apperr.KindConfig, "remote.max_connections must not exceed %d", MaxRemoteConnections)
	}
	return nil
}

const originForm = "must be an exact lowercase http(s)://host[:port] origin"

// portProblem returns a non-empty reason if p is not a canonical decimal
// port (1-65535, no sign, no leading zeros).
func portProblem(p string) string {
	if p == "" {
		return "empty port"
	}
	for _, r := range p {
		if r < '0' || r > '9' {
			return "port must be 1-65535"
		}
	}
	if len(p) > 1 && p[0] == '0' {
		return "port must not have leading zeros"
	}
	if n, err := strconv.Atoi(p); err != nil || n < 1 || n > 65535 {
		return "port must be 1-65535"
	}
	return ""
}

// ipv6Problem returns a non-empty reason if s (without brackets) is not a
// canonical lowercase IPv6 literal.
func ipv6Problem(s string) string {
	ip := net.ParseIP(s)
	if ip == nil || !strings.Contains(s, ":") {
		return "must contain a valid IPv6 literal"
	}
	if ip.String() != s {
		return "IPv6 literal must be in canonical compressed form"
	}
	return ""
}

// hostNameProblem validates a DNS name or IPv4 literal (no port): lowercase
// ASCII letters, digits, '.', '-', at most 253 bytes, no empty label, no
// trailing dot, labels of at most 63 bytes that neither start nor end with
// a hyphen.
func hostNameProblem(h string) string {
	if len(h) > 253 {
		return "host exceeds 253 characters"
	}
	if strings.HasSuffix(h, ".") {
		return "trailing dot is not allowed"
	}
	for _, r := range h {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '.' && r != '-' {
			return hostForm
		}
	}
	for _, label := range strings.Split(h, ".") {
		switch {
		case label == "":
			return "host has an empty label"
		case len(label) > 63:
			return "host label exceeds 63 characters"
		case label[0] == '-' || label[len(label)-1] == '-':
			return "host label must not start or end with a hyphen"
		}
	}
	if r := numericHostProblem(h); r != "" {
		return r
	}
	return ""
}

// numericHostProblem rejects numeric-looking hosts that are not a canonical
// dotted-quad IPv4 literal. A host whose last label is all digits (or that
// consists only of digits and dots) is parsed by many resolvers as an
// abbreviated, octal, or integer IPv4 form (127.1, 2130706433, 01.2.3.4), so
// only the exact canonical form is accepted.
func numericHostProblem(h string) string {
	last := h[strings.LastIndexByte(h, '.')+1:]
	numeric := last != "" && strings.Trim(last, "0123456789") == ""
	if !numeric && strings.Trim(h, "0123456789.") != "" {
		return ""
	}
	if ip, err := netip.ParseAddr(h); err == nil && ip.Is4() && ip.String() == h {
		return ""
	}
	return "host is a non-canonical IPv4 address (use dotted-quad without leading zeros)"
}

const hostForm = "must be a host[:port] (no scheme, path, userinfo, or whitespace)"

// originProblem returns a non-empty reason if o is not an exact origin in
// canonical form (ADR MVP-2: default port omitted). The reason never
// includes o itself (it may carry userinfo).
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
	if strings.Contains(o, "%") {
		return originForm + " (percent-encoding is not allowed)"
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
	host := u.Hostname()
	if host == "" {
		return originForm + " (host is required)"
	}
	if p := u.Port(); p != "" {
		if r := portProblem(p); r != "" {
			return originForm + " (" + r + ")"
		}
		if (u.Scheme == "https" && p == "443") || (u.Scheme == "http" && p == "80") {
			return originForm + " (default port must be omitted)"
		}
	} else if strings.HasSuffix(u.Host, ":") {
		return originForm + " (empty port)"
	}
	if strings.Contains(u.Host, "[") && !strings.Contains(host, ":") {
		return originForm + " (brackets are only allowed around an IPv6 literal)"
	}
	if strings.Contains(host, ":") {
		if r := ipv6Problem(host); r != "" {
			return originForm + " (" + r + ")"
		}
	} else if hostNameProblem(host) != "" {
		return originForm + " (invalid host)"
	}
	return ""
}

// hostProblem returns a non-empty reason if h is not a canonical
// allowed_hosts entry: a lowercase host name, IPv4 literal, or bracketed
// IPv6 literal with an optional :port. An entry without a port matches the
// default port only (the matcher is internal/transport hostorigin.go).
func hostProblem(h string) string {
	if h == "" {
		return "must not be empty"
	}
	if strings.Contains(h, "*") {
		return "must not contain a wildcard"
	}
	if h != strings.ToLower(h) {
		return "must be a lowercase host (uppercase is not allowed)"
	}
	for _, r := range h {
		if r <= ' ' || r == 0x7f {
			return "whitespace/control characters are not allowed"
		}
		if r > 0x7e {
			return "must be ASCII (use the punycode form)"
		}
	}
	if strings.HasPrefix(h, "[") {
		end := strings.LastIndex(h, "]")
		if end < 0 {
			return hostForm
		}
		if r := ipv6Problem(h[1:end]); r != "" {
			return r
		}
		switch rest := h[end+1:]; {
		case rest == "":
			return ""
		case rest[0] == ':':
			return portProblem(rest[1:])
		default:
			return hostForm
		}
	}
	if strings.ContainsAny(h, `/\@?#%[]`) {
		return hostForm
	}
	name := h
	switch strings.Count(h, ":") {
	case 0:
	case 1:
		i := strings.IndexByte(h, ':')
		name = h[:i]
		if r := portProblem(h[i+1:]); r != "" {
			return r
		}
	default:
		return hostForm
	}
	return hostNameProblem(name)
}
