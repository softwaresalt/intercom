package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestNumericHostsMustBeCanonicalIPv4(t *testing.T) {
	bad := []string{"999.999.999.999", "127.000.000.001", "127.1", "2130706433", "01.2.3.4", "1.2.3", "1.2.3.4.5", "app.1.2.3.4x.99"}
	for _, h := range bad[:7] {
		for _, entry := range []string{h, h + ":3000"} {
			cfg := validRemoteCfg(t)
			cfg.Remote.AllowedHosts = []string{entry}
			_, err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), "non-canonical IPv4 address") {
				t.Errorf("host %q: err = %v, want non-canonical IPv4 address", entry, err)
			}
		}
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = []string{"http://" + h}
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[0]") {
			t.Errorf("origin host %q: err = %v, want rejection", h, err)
		}
	}
	for _, h := range []string{"127.0.0.1", "10.20.30.40:8080", "a1.example.com", "example.123abc", "1password.example"} {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedHosts = []string{h}
		if _, err := cfg.Validate(); err != nil {
			t.Errorf("host %q: unexpected error %v", h, err)
		}
	}
}

func TestBracketedNonIPv6OriginsRejected(t *testing.T) {
	for _, o := range []string{"https://[example.com]", "http://[127.0.0.1]:3000", "http://[localhost]", "https://[1.2.3.4]"} {
		cfg := validRemoteCfg(t)
		cfg.Remote.AllowedOrigins = []string{o}
		_, err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "remote.allowed_origins[0]") {
			t.Errorf("origin %q: err = %v, want rejection", o, err)
		}
	}
}

func TestDecodeSyntaxErrorNeverEchoesValueFragments(t *testing.T) {
	docs := []struct {
		name, doc string
		secrets   []string
	}{
		{"unquoted", "default_workspace_root = \".\"\n[remote]\npassword = hunter2\n", []string{"hunter2", "hunt"}},
		{"dotted", "remote.password = hunter2\n", []string{"hunter2", "hunt"}},
		{"unterminated", "default_workspace_root = \"hunter2\n", []string{"hunter2"}},
		{"badArray", "default_workspace_root = \".\"\nx = [\"hunter2\" \"zz9plural\"]\n", []string{"hunter2", "zz9plural"}},
		{"doubleEq", "x = = hunter2\n", []string{"hunter2"}},
	}
	for _, tc := range docs {
		_, _, err := Decode(tc.doc)
		if err == nil {
			t.Errorf("%s: expected error", tc.name)
			continue
		}
		var ae *apperr.Error
		if !errors.As(err, &ae) || ae.Kind() != apperr.KindConfig {
			t.Errorf("%s: not KindConfig apperr: %v", tc.name, err)
		}
		msg := err.Error()
		for _, s := range tc.secrets {
			if strings.Contains(msg, s) {
				t.Errorf("%s: message echoes %q: %s", tc.name, s, msg)
			}
		}
		if !strings.Contains(msg, "line ") {
			t.Errorf("%s: message lacks line number: %s", tc.name, msg)
		}
	}
}
