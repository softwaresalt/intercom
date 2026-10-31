package config

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

func validRemoteCfg(t *testing.T) *Config {
	t.Helper()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	return cfg
}

func tomlRoot(t *testing.T) string {
	t.Helper()
	return "default_workspace_root = \"" + strings.ReplaceAll(t.TempDir(), `\`, `/`) + "\"\n"
}

func TestDefaultRemoteIsSecure(t *testing.T) {
	r := Default().Remote
	if r.TokenLifetimeSeconds != 14400 {
		t.Errorf("TokenLifetimeSeconds = %d, want 14400", r.TokenLifetimeSeconds)
	}
	if r.MaxConnections != DefaultRemoteMaxConnections {
		t.Errorf("MaxConnections = %d, want %d", r.MaxConnections, DefaultRemoteMaxConnections)
	}
	if DefaultRemoteMaxConnections != hub.DefaultMaxRemote {
		t.Errorf("DefaultRemoteMaxConnections = %d, want hub.DefaultMaxRemote %d", DefaultRemoteMaxConnections, hub.DefaultMaxRemote)
	}
	if MaxRemoteConnections != hub.MaxRemoteHardLimit {
		t.Errorf("MaxRemoteConnections = %d, want hub.MaxRemoteHardLimit %d", MaxRemoteConnections, hub.MaxRemoteHardLimit)
	}
	if len(r.AllowedOrigins) != 0 || len(r.AllowedHosts) != 0 {
		t.Errorf("origins/hosts = %v/%v, want empty", r.AllowedOrigins, r.AllowedHosts)
	}
	if r.RemoteApproveDestructive {
		t.Error("RemoteApproveDestructive default = true, want false")
	}
	if _, err := validRemoteCfg(t).Validate(); err != nil {
		t.Fatalf("Default remote must validate: %v", err)
	}
}

func TestValidateRemoteTokenLifetimeBoundaries(t *testing.T) {
	const msg = "remote.token_lifetime_seconds must be between 60 and 86400"
	for _, tc := range []struct {
		v    int
		fail bool
	}{{-1, true}, {0, true}, {59, true}, {60, false}, {14400, false}, {86400, false}, {86401, true}} {
		cfg := validRemoteCfg(t)
		cfg.Remote.TokenLifetimeSeconds = tc.v
		_, err := cfg.Validate()
		if !tc.fail {
			if err != nil {
				t.Errorf("lifetime %d: unexpected error %v", tc.v, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), msg) {
			t.Errorf("lifetime %d: error = %v, want containing %q", tc.v, err, msg)
			continue
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind() != apperr.KindConfig {
			t.Errorf("lifetime %d: want apperr.KindConfig, got %v", tc.v, err)
		}
	}
}

func TestValidateRemoteMaxConnections(t *testing.T) {
	const (
		low  = "remote.max_connections must be at least 1"
		high = "remote.max_connections must not exceed 1024"
	)
	if MaxRemoteConnections != 1024 {
		t.Fatalf("MaxRemoteConnections = %d, want 1024", MaxRemoteConnections)
	}
	for _, tc := range []struct {
		v    int
		want string
	}{{-1, low}, {0, low}, {1, ""}, {8, ""}, {1024, ""}, {1025, high}} {
		cfg := validRemoteCfg(t)
		cfg.Remote.MaxConnections = tc.v
		_, err := cfg.Validate()
		if tc.want == "" {
			if err != nil {
				t.Errorf("max_connections %d: unexpected error %v", tc.v, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("max_connections %d: err = %v, want %q", tc.v, err, tc.want)
		}
	}
}
func TestValidateRemoteStructLiteralZeroFailsClosed(t *testing.T) {
	cfg := &Config{
		MaxConcurrentSessions: 1,
		DefaultWorkspaceRoot:  t.TempDir(),
		Copilot:               CopilotConfig{StartupTimeoutSeconds: 30},
	}
	_, err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "remote.token_lifetime_seconds must be between 60 and 86400") {
		t.Fatalf("zero-value remote: err = %v", err)
	}
}

func TestValidateRemoteOrigins(t *testing.T) {
	good := [][]string{
		nil,
		{"https://app.example.com"},
		{"http://localhost:5173", "https://app.example.com:8443"},
		{"http://127.0.0.1:3000"},
		{"http://[::1]:3000"},
		{"https://[::1]"},
		{"https://[2001:db8::1]:8443"},
		{"https://example.com:444", "http://example.com:443", "https://example.com:80"},
	}
	for _, o := range good {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = o
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("origins %v: unexpected error %v", o, err)
		}
	}
	const form = "must be an exact lowercase http(s)://host[:port] origin"
	bad := []struct {
		name, entry, reason string
	}{
		{"empty", "", "must not be empty"},
		{"wildcard", "*", "must not contain a wildcard"},
		{"wildcardHost", "https://*.example.com", "must not contain a wildcard"},
		{"wildcardPort", "https://example.com:*", "must not contain a wildcard"},
		{"path", "https://example.com/app", "no path, query, or fragment"},
		{"trailing", "https://example.com/", "no path, query, or fragment"},
		{"query", "https://example.com?x=1", "no path, query, or fragment"},
		{"forceQuery", "https://example.com?", "no path, query, or fragment"},
		{"fragment", "https://example.com#x", "no path, query, or fragment"},
		{"userinfo", "https://user:hunter2@example.com", "userinfo is not allowed"},
		{"userinfoNoPass", "https://hunter2@example.com", "userinfo is not allowed"},
		{"uppercase", "https://Example.com", "uppercase is not allowed"},
		{"upperScheme", "HTTPS://example.com", "uppercase is not allowed"},
		{"noScheme", "example.com", "scheme must be http or https"},
		{"badScheme", "ftp://example.com", "scheme must be http or https"},
		{"noHost", "https://", "host is required"},
		{"badPort", "https://example.com:99999", "port must be 1-65535"},
		{"zeroPort", "https://example.com:0", "port must be 1-65535"},
		{"emptyPort", "https://example.com:", "empty port"},
		{"leadingZeroPort", "https://example.com:08443", "leading zeros"},
		{"defaultHTTPS", "https://example.com:443", "default port must be omitted"},
		{"defaultHTTP", "http://example.com:80", "default port must be omitted"},
		{"percent", "https://exa%6dple.com", "percent-encoding is not allowed"},
		{"space", "https://exa mple.com", "whitespace/control characters are not allowed"},
		{"trailingDot", "https://example.com.", "invalid host"},
		{"ipv6Upper", "http://[::FFFF]:3000", "uppercase is not allowed"},
		{"ipv6Unclosed", "http://[::1", form},
		{"ipv6EmptyPort", "http://[::1]:", "empty port"},
		{"ipv6DefaultPort", "https://[::1]:443", "default port must be omitted"},
		{"ipv6NonCanonical", "http://[0:0:0:0:0:0:0:1]", "canonical compressed form"},
		{"ipv6Zone", "http://[fe80::1%25eth0]", "percent-encoding is not allowed"},
	}
	for _, tc := range bad {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = []string{tc.entry}
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[0] ") || !strings.Contains(err.Error(), tc.reason) {
			t.Errorf("%s: error = %v, want remote.allowed_origins[0] ... %q", tc.name, err, tc.reason)
			continue
		}
		if strings.Contains(err.Error(), "hunter2") {
			t.Errorf("%s: error echoes userinfo secret: %v", tc.name, err)
		}
	}
}

func TestValidateRemoteOriginDuplicateMessage(t *testing.T) {
	cfg := validRemoteCfg(t)
	cfg.Remote.AllowedOrigins = []string{"https://a.example.com", "http://[::1]:3000", "http://[::1]:3000"}
	_, err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[2] is a duplicate entry") {
		t.Errorf("err = %v", err)
	}
}

func TestValidateRemoteHosts(t *testing.T) {
	good := [][]string{
		nil, {"app.example.com"}, {"localhost", "127.0.0.1"},
		{"localhost:3000", "127.0.0.1:8080", "app.example.com:443"},
		{"[::1]", "[::1]:3000", "[2001:db8::1]:8443"},
		{"xn--bcher-kva.example", "a-b.example"},
		{strings.Repeat("a", 63) + "." + strings.Repeat("b", 63) + "." + strings.Repeat("c", 63) + "." + strings.Repeat("d", 61)},
	}
	for _, h := range good {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = h
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("hosts %v: unexpected error %v", h, err)
		}
	}
	const plain = "must be a host[:port] (no scheme, path, userinfo, or whitespace)"
	bad := []struct {
		name, entry, reason string
	}{
		{"empty", "", "must not be empty"},
		{"wildcard", "*", "must not contain a wildcard"},
		{"wildSub", "*.example.com", "must not contain a wildcard"},
		{"wildPort", "example.com:*", "must not contain a wildcard"},
		{"scheme", "https://example.com", plain},
		{"path", "example.com/x", plain},
		{"userinfo", "u@example.com", plain},
		{"uppercase", "Example.com", "uppercase is not allowed"},
		{"space", "exa mple.com", "whitespace/control characters are not allowed"},
		{"control", "example.com\x00", "whitespace/control characters are not allowed"},
		{"nonASCII", "b\u00fccher.example", "must be ASCII"},
		{"trailingDot", "example.com.", "trailing dot is not allowed"},
		{"leadingDot", ".example.com", "empty label"},
		{"emptyLabel", "a..example.com", "empty label"},
		{"tooLong", strings.Repeat("a.", 127) + "a", "exceeds 253"},
		{"longLabel", strings.Repeat("a", 64) + ".example", "label exceeds 63"},
		{"emptyPort", "example.com:", "empty port"},
		{"zeroPort", "example.com:0", "port must be 1-65535"},
		{"bigPort", "example.com:65536", "port must be 1-65535"},
		{"leadingZeroPort", "example.com:080", "leading zeros"},
		{"signPort", "example.com:+80", "port must be 1-65535"},
		{"nonNumericPort", "example.com:http", "port must be 1-65535"},
		{"twoPorts", "example.com:80:81", plain},
		{"bareIPv6", "::1", plain},
		{"ipv6Unclosed", "[::1", plain},
		{"ipv6EmptyPort", "[::1]:", "empty port"},
		{"ipv6Junk", "[::1]x", plain},
		{"ipv6NonCanonical", "[0:0:0:0:0:0:0:1]", "canonical compressed form"},
		{"ipv6Upper", "[::FFFF]", "uppercase is not allowed"},
		{"ipv6NotIP", "[example]", "valid IPv6 literal"},
		{"ipv6Zone", "[fe80::1%eth0]", "valid IPv6 literal"},
		{"hyphenLabel", "-a.example", "must not start or end with a hyphen"},
		{"duplicate", "a.example.com", "is a duplicate entry"},
	}
	for _, tc := range bad {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = []string{tc.entry}
		if tc.name == "duplicate" {
			cfg.Remote.AllowedHosts = []string{"x.example.com", tc.entry, tc.entry}
		}
		_, err := cfg.Validate()
		prefix := "remote.allowed_hosts[0] "
		if tc.name == "duplicate" {
			prefix = "remote.allowed_hosts[2] "
		}
		if err == nil || !strings.Contains(err.Error(), prefix) || !strings.Contains(err.Error(), tc.reason) {
			t.Errorf("%s: error = %v, want %s... %q", tc.name, err, prefix, tc.reason)
		}
	}
}

func TestValidateRemoteIntraGroupOrder(t *testing.T) {
	cfg := validRemoteCfg(t)
	cfg.Remote.TokenLifetimeSeconds = 1
	cfg.Remote.AllowedOrigins = []string{"*"}
	cfg.Remote.AllowedHosts = []string{"*"}
	cfg.Remote.MaxConnections = 0
	want := []string{"remote.token_lifetime_seconds", "remote.allowed_origins[0]", "remote.allowed_hosts[0]", "remote.max_connections"}
	for i, w := range want {
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), w) {
			t.Fatalf("step %d: err = %v, want %s", i, err, w)
		}
		switch i {
		case 0:
			cfg.Remote.TokenLifetimeSeconds = 60
		case 1:
			cfg.Remote.AllowedOrigins = nil
		case 2:
			cfg.Remote.AllowedHosts = nil
		}
	}
}
func TestValidateRemoteRunsAfterShippedRulesAndNeverMutates(t *testing.T) {
	// An earlier rule (database.path) wins over a remote violation.
	cfg := validRemoteCfg(t)
	cfg.Database.Path = "../x.db"
	cfg.Remote.MaxConnections = 0
	if _, err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "database.path") {
		t.Errorf("rule order: err = %v, want database.path rule first", err)
	}

	// A remote failure leaves the receiver untouched.
	cfg = validRemoteCfg(t)
	original := cfg.DefaultWorkspaceRoot
	cfg.DefaultWorkspaceRoot = original + string(os.PathSeparator) + "."
	before := cfg.DefaultWorkspaceRoot
	cfg.Remote.MaxConnections = 0
	if _, err := cfg.Validate(); err == nil {
		t.Fatal("want remote error")
	}
	if cfg.DefaultWorkspaceRoot != before {
		t.Errorf("DefaultWorkspaceRoot mutated on failed validation: %q -> %q", before, cfg.DefaultWorkspaceRoot)
	}
}

func TestDecodeRemoteSection(t *testing.T) {
	cfg, report, err := Decode(tomlRoot(t) + `[remote]
token_lifetime_seconds = 3600
allowed_origins = ["https://app.example.com"]
allowed_hosts = ["app.example.com"]
max_connections = 2
remote_approve_destructive = true
`)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(report.UnknownKeys) != 0 {
		t.Errorf("known remote keys reported unknown: %v", report.UnknownKeys)
	}
	r := cfg.Remote
	if r.TokenLifetimeSeconds != 3600 || r.MaxConnections != 2 || !r.RemoteApproveDestructive ||
		!slices.Equal(r.AllowedOrigins, []string{"https://app.example.com"}) ||
		!slices.Equal(r.AllowedHosts, []string{"app.example.com"}) {
		t.Errorf("decoded remote = %+v", r)
	}
}

func TestDecodeRemoteAbsentUsesSecureDefaults(t *testing.T) {
	cfg, _, err := Decode(tomlRoot(t))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if cfg.Remote.TokenLifetimeSeconds != 14400 || cfg.Remote.MaxConnections != 8 || cfg.Remote.RemoteApproveDestructive {
		t.Errorf("remote defaults = %+v", cfg.Remote)
	}
}

func TestDecodeRemoteUnknownKeyStillReported(t *testing.T) {
	_, report, err := Decode(tomlRoot(t) + "[remote]\nmax_connections = 3\nfuture_option = 1\n")
	if err == nil {
		t.Fatal("Decode accepted an unknown key under [remote]")
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind() != apperr.KindConfig {
		t.Fatalf("want KindConfig, got %v", err)
	}
	if !strings.Contains(err.Error(), "remote.future_option is an unknown key under [remote]; only known [remote] schema keys are permitted") {
		t.Errorf("err = %v", err)
	}
	if !slices.Equal(report.UnknownKeys, []string{"remote.future_option"}) {
		t.Errorf("UnknownKeys = %v, want [remote.future_option]", report.UnknownKeys)
	}

	_, _, err = Decode(tomlRoot(t) + "[remote.sub]\nsetting = \"v-9f8e\"\n")
	if err == nil || !strings.Contains(err.Error(), "is an unknown key under [remote]") {
		t.Errorf("nested unknown table: err = %v", err)
	}
	if err != nil && strings.Contains(err.Error(), "v-9f8e") {
		t.Errorf("error echoes value: %v", err)
	}

	// The credential-shaped error takes precedence over the unknown-key one.
	_, _, err = Decode(tomlRoot(t) + "[remote]\nfuture_option = 1\ntoken = \"secret\"\n")
	if err == nil || !strings.Contains(err.Error(), "credential-shaped") {
		t.Errorf("precedence: err = %v, want credential-shaped error", err)
	}
}

func TestValidateRemoteRejectsHexIPv4HostsAndOrigins(t *testing.T) {
	hosts := []string{"0x7f.0x1", "0x7f.0.0.0x1", "127.0.0.0x1", "0x7f000001", "0x0"}
	for _, h := range hosts {
		for _, entry := range []string{h, h + ":3000"} {
			cfg := validRemoteCfg(t)
			cfg.Remote.AllowedHosts = []string{entry}
			_, err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), "non-canonical IPv4 address") {
				t.Errorf("allowed_hosts %q: err = %v, want non-canonical IPv4 rejection", entry, err)
			}
		}
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = []string{"http://" + h}
		_, err := cfg.Validate()
		if err == nil {
			t.Errorf("allowed_origins http://%s accepted, want rejection", h)
		}
	}
	for _, h := range []string{"1password.example", "example.123abc", "a1.example.com"} {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = []string{h}
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("allowed_hosts %q rejected: %v", h, err)
		}
	}
}

func TestDecodeRemoteInvalidValuesFieldSpecific(t *testing.T) {
	_, _, err := Decode(tomlRoot(t) + "[remote]\ntoken_lifetime_seconds = 59\n")
	if err == nil || !strings.Contains(err.Error(), "remote.token_lifetime_seconds must be between 60 and 86400") {
		t.Errorf("err = %v", err)
	}
	_, _, err = Decode(tomlRoot(t) + "[remote]\nmax_connections = 0\n")
	if err == nil || !strings.Contains(err.Error(), "remote.max_connections must be at least 1") {
		t.Errorf("err = %v", err)
	}
	_, _, err = Decode(tomlRoot(t) + "[remote]\nallowed_origins = [\"*\"]\n")
	if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[0]") {
		t.Errorf("err = %v", err)
	}
}

func TestDecodeRemoteCredentialShapedKeyRejectedWithoutEchoingValue(t *testing.T) {
	const secret = "s3cr3t-VALUE-9f8e7d"
	keys := []string{"token", "secret", "password", "api_key", "apiKey", "credential", "credentials",
		"bearer", "bearer_token", "auth_token", "private_key", "access_key", "Token", "PASSWORD", `"my-secret"`, "jwt", "client_secret",
		"pass", "passcode", "passphrase", "pat", "psk", "otp", "pwd", "cookie", "signature", "session", "hmac", "salt", "cert", "dsn", "connstr", "connection_string", "pin", "github_pat"}
	for _, k := range keys {
		cfg, report, err := Decode(tomlRoot(t) + "[remote]\n" + k + " = \"" + secret + "\"\n")
		_ = cfg
		if err == nil {
			t.Errorf("key %s: accepted, want rejection", k)
			continue
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind() != apperr.KindConfig {
			t.Errorf("key %s: want KindConfig, got %v", k, err)
		}
		bare := strings.Trim(k, `"`)
		if !strings.Contains(err.Error(), "remote."+bare) {
			t.Errorf("key %s: error %q does not name the key", k, err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("key %s: error echoes value: %v", k, err)
		}
		for _, u := range report.UnknownKeys {
			if strings.Contains(u, secret) {
				t.Errorf("key %s: UnknownKeys echoes value: %v", k, u)
			}
		}
		for _, w := range report.Warnings {
			if strings.Contains(w, secret) {
				t.Errorf("key %s: Warnings echoes value: %v", k, w)
			}
		}
	}
}

func TestDecodeRemoteCredentialShapedNestedTableAndNonStringRejected(t *testing.T) {
	const secret = "nested-VALUE-123"
	for _, doc := range []string{
		"[remote.auth]\nvalue = \"" + secret + "\"\n",
		"[remote]\nsub = { password = \"" + secret + "\" }\n",
		"[remote]\ntoken = 12345\n",
		"[remote]\nsecret = [\"" + secret + "\"]\n",
	} {
		_, _, err := Decode(tomlRoot(t) + doc)
		if err == nil {
			t.Errorf("doc %q: accepted", doc)
			continue
		}
		if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "12345") {
			t.Errorf("doc %q: error echoes value: %v", doc, err)
		}
	}
}

func TestDecodeKnownRemoteKeysAreNotCredentialShaped(t *testing.T) {
	// token_lifetime_seconds contains "token" but is a known, non-secret field.
	if _, _, err := Decode(tomlRoot(t) + "[remote]\ntoken_lifetime_seconds = 120\n"); err != nil {
		t.Fatalf("token_lifetime_seconds rejected as credential: %v", err)
	}
	// Credential-looking keys outside [remote] keep tolerant behavior.
	_, report, err := Decode(tomlRoot(t) + "token = \"x\"\n")
	if err != nil {
		t.Fatalf("top-level unknown key must stay tolerant: %v", err)
	}
	if !slices.Equal(report.UnknownKeys, []string{"token"}) {
		t.Errorf("UnknownKeys = %v", report.UnknownKeys)
	}
}

func TestExampleConfigDocumentsRemote(t *testing.T) {
	data, err := os.ReadFile(exampleConfigPath(t))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	if !strings.Contains(doc, "[remote]") {
		t.Fatal("config.toml.example does not document [remote]")
	}
	for _, k := range []string{"token_lifetime_seconds", "allowed_origins", "allowed_hosts", "max_connections", "remote_approve_destructive"} {
		if !strings.Contains(doc, k) {
			t.Errorf("config.toml.example missing %s", k)
		}
	}
	cfg, _, err := Load(exampleConfigPath(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Remote.AllowedOrigins) != 0 || len(cfg.Remote.AllowedHosts) != 0 || cfg.Remote.RemoteApproveDestructive {
		t.Errorf("example remote must have secure defaults: %+v", cfg.Remote)
	}
	// No credential-shaped assignment lines inside [remote] in the example.
	inRemote := false
	for _, line := range strings.Split(doc, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "[") {
			inRemote = strings.HasPrefix(l, "[remote")
			continue
		}
		if !inRemote || strings.HasPrefix(l, "#") || !strings.Contains(l, "=") {
			continue
		}
		if k := strings.ToLower(strings.TrimSpace(strings.SplitN(l, "=", 2)[0])); isCredentialShapedKey(k) && !isKnownRemoteKey(k) {
			t.Errorf("config.toml.example contains credential-shaped key %q", k)
		}
	}
}

func TestConfigReferenceDocumentsRemote(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatal(err)
	}
	doc := string(data)
	for _, want := range []string{
		"[remote]", "Remote.TokenLifetimeSeconds", "Remote.MaxConnections", "Remote.AllowedOrigins",
		"Remote.AllowedHosts", "Remote.RemoteApproveDestructive",
		"remote.token_lifetime_seconds must be between 60 and 86400",
		"remote.max_connections must be at least 1", "remote.max_connections must not exceed 1024", "MaxRemoteConnections", "default port", "absent `[remote]` section",
		"remote.allowed_origins[{i}]", "remote.allowed_hosts[{i}]",
		"credential-shaped",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("docs/config-reference.md missing %q", want)
		}
	}
}

func TestDecodeRemoteOrderCredentialKeyBeatsMissingRoot(t *testing.T) {
	// Pinned current behaviour: the credential screen runs before the
	// required-key check, so a credential key wins over a missing root.
	_, _, err := Decode("[remote]\ntoken = \"x\"\n")
	if err == nil || !strings.Contains(err.Error(), "credential-shaped") {
		t.Fatalf("err = %v, want credential-shaped key error", err)
	}
}

func TestDecodeRemoteNonBooleanApproveDestructiveRejected(t *testing.T) {
	for _, v := range []string{`"true"`, "1", "[true]"} {
		_, _, err := Decode(tomlRoot(t) + "[remote]\nremote_approve_destructive = " + v + "\n")
		if err == nil {
			t.Errorf("value %s accepted, want type error", v)
			continue
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind() != apperr.KindConfig {
			t.Errorf("value %s: want KindConfig, got %v", v, err)
		}
	}
}

func TestDecodeRemoteMaxConnectionsUpperBound(t *testing.T) {
	_, _, err := Decode(tomlRoot(t) + "[remote]\nmax_connections = 1025\n")
	if err == nil || !strings.Contains(err.Error(), "remote.max_connections must not exceed 1024") {
		t.Errorf("err = %v", err)
	}
}
