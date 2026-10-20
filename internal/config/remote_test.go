package config

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
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
	if r.MaxConnections != 8 {
		t.Errorf("MaxConnections = %d, want 8", r.MaxConnections)
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
	const msg = "remote.max_connections must be at least 1"
	for _, tc := range []struct {
		v    int
		fail bool
	}{{-1, true}, {0, true}, {1, false}, {8, false}} {
		cfg := validRemoteCfg(t)
		cfg.Remote.MaxConnections = tc.v
		_, err := cfg.Validate()
		if tc.fail != (err != nil) || (tc.fail && !strings.Contains(err.Error(), msg)) {
			t.Errorf("max_connections %d: err = %v, fail want %v (%q)", tc.v, err, tc.fail, msg)
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
	}
	for _, o := range good {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = o
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("origins %v: unexpected error %v", o, err)
		}
	}
	bad := map[string][]string{
		"empty":        {""},
		"wildcard":     {"*"},
		"wildcardHost": {"https://*.example.com"},
		"wildcardPort": {"https://example.com:*"},
		"path":         {"https://example.com/app"},
		"trailing":     {"https://example.com/"},
		"query":        {"https://example.com?x=1"},
		"fragment":     {"https://example.com#x"},
		"userinfo":     {"https://user:hunter2@example.com"},
		"uppercase":    {"https://Example.com"},
		"upperScheme":  {"HTTPS://example.com"},
		"noScheme":     {"example.com"},
		"badScheme":    {"ftp://example.com"},
		"noHost":       {"https://"},
		"badPort":      {"https://example.com:99999"},
		"space":        {"https://exa mple.com"},
		"duplicate":    {"https://a.example.com", "https://a.example.com"},
	}
	for name, o := range bad {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = o
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[") {
			t.Errorf("%s: error = %v, want field-specific remote.allowed_origins[i]", name, err)
			continue
		}
		if strings.Contains(err.Error(), "hunter2") {
			t.Errorf("%s: error echoes userinfo secret: %v", name, err)
		}
	}
}

func TestValidateRemoteHosts(t *testing.T) {
	for _, h := range [][]string{nil, {"app.example.com"}, {"localhost", "127.0.0.1"}} {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = h
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("hosts %v: unexpected error %v", h, err)
		}
	}
	bad := map[string][]string{
		"empty":     {""},
		"wildcard":  {"*"},
		"wildSub":   {"*.example.com"},
		"scheme":    {"https://example.com"},
		"path":      {"example.com/x"},
		"port":      {"example.com:80"},
		"userinfo":  {"u@example.com"},
		"uppercase": {"Example.com"},
		"space":     {"exa mple.com"},
		"duplicate": {"a.example.com", "a.example.com"},
	}
	for name, h := range bad {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = h
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "remote.allowed_hosts[") {
			t.Errorf("%s: error = %v, want field-specific remote.allowed_hosts[i]", name, err)
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
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !slices.Equal(report.UnknownKeys, []string{"remote.future_option"}) {
		t.Errorf("UnknownKeys = %v, want [remote.future_option]", report.UnknownKeys)
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
		"bearer", "bearer_token", "auth_token", "private_key", "access_key", "Token", "PASSWORD", `"my-secret"`, "jwt", "client_secret"}
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
		if k := strings.ToLower(strings.TrimSpace(strings.SplitN(l, "=", 2)[0])); isCredentialShapedKey(k) && k != "token_lifetime_seconds" {
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
		"remote.max_connections must be at least 1",
		"remote.allowed_origins[{i}]", "remote.allowed_hosts[{i}]",
		"credential-shaped",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("docs/config-reference.md missing %q", want)
		}
	}
}
