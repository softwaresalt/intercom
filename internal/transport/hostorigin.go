package transport

import (
	"net/netip"
	"strings"
)

// canonHostPort canonicalises a host[:port] (the HTTP Host header, an
// allowed_hosts entry, or the authority of an Origin): lowercase ASCII, a
// trailing dot stripped from a name, IPv6 literals bracketed and compressed,
// port digits only. It returns the host part (brackets kept for IPv6), the
// port ("" if absent) and whether the input was well-formed. Anything with
// userinfo, whitespace, control or non-ASCII characters, a path, or an
// unbracketed IPv6 literal is malformed (fail closed; IDN hosts must arrive
// in punycode form, which is ASCII).
func canonHostPort(s string) (host, port string, ok bool) {
	if s == "" || len(s) > 261 {
		return "", "", false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c <= ' ' || c >= 0x7f || strings.IndexByte(`/\@?#%*`, c) >= 0 {
			return "", "", false
		}
	}
	s = strings.ToLower(s)
	rest := ""
	if strings.HasPrefix(s, "[") {
		end := strings.IndexByte(s, ']')
		if end < 0 {
			return "", "", false
		}
		ip, err := netip.ParseAddr(s[1:end])
		if err != nil || !ip.Is6() || ip.Zone() != "" {
			return "", "", false
		}
		host, rest = "["+ip.String()+"]", s[end+1:]
	} else {
		name := s
		if i := strings.IndexByte(s, ':'); i >= 0 {
			name, rest = s[:i], s[i:]
			if strings.IndexByte(rest[1:], ':') >= 0 {
				return "", "", false
			}
		}
		name = strings.TrimSuffix(name, ".")
		if name == "" || strings.HasPrefix(name, ".") || strings.Contains(name, "..") {
			return "", "", false
		}
		host = name
	}
	switch {
	case rest == "":
	case rest[0] == ':' && len(rest) > 1 && len(rest) <= 6:
		for _, c := range rest[1:] {
			if c < '0' || c > '9' {
				return "", "", false
			}
		}
		port = strings.TrimLeft(rest[1:], "0")
		if port == "" || len(port) > 5 || (len(port) == 5 && port > "65535") {
			return "", "", false
		}
	default:
		return "", "", false
	}
	return host, port, true
}

// hostKey is the canonical comparison key for a Host header or an
// allowed_hosts entry: host, plus ":port" when a port is present. An entry
// without a port therefore matches only a Host without a port (the default
// port), per ADR MVP-2.
func hostKey(s string) (string, bool) {
	host, port, ok := canonHostPort(s)
	if !ok {
		return "", false
	}
	if port != "" {
		return host + ":" + port, true
	}
	return host, true
}

// originKey is the canonical comparison key for an Origin header or an
// allowed_origins entry: scheme://host[:port] with scheme/host lowercase, a
// trailing host dot stripped and the scheme's default port omitted. The
// literal "null", paths, queries and non-http(s) schemes are malformed.
func originKey(s string) (string, bool) {
	scheme, authority, found := strings.Cut(s, "://")
	if !found {
		return "", false
	}
	scheme = strings.ToLower(scheme)
	if scheme != "http" && scheme != "https" {
		return "", false
	}
	host, port, ok := canonHostPort(authority)
	if !ok {
		return "", false
	}
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		port = ""
	}
	if port != "" {
		host += ":" + port
	}
	return scheme + "://" + host, true
}
