package main

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/softwaresalt/intercom-go/internal/config"
)

const runbookPath = "../../docs/runbooks/intercom-operator-runbook.md"

// invocation matches a command line that starts intercom (installed binary or
// `go run`), at the start of a line or of an inline code span.
var invocation = regexp.MustCompile("(?:^[ \\t]*|`|\\$ )(?:go run \\./cmd/intercom|intercom(?:\\.exe)?)[ \\t]+(.*)$")

var flagToken = regexp.MustCompile(`--[a-z][a-z0-9-]*`)

func runbook(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(runbookPath)
	if err != nil {
		t.Fatalf("read the operator runbook: %v", err)
	}
	return string(b)
}

func realFlags(t *testing.T) map[string]bool {
	t.Helper()
	cmd := newRootCmd()
	cmd.InitDefaultHelpFlag()
	flags := map[string]bool{}
	cmd.Flags().VisitAll(func(f *pflag.Flag) { flags["--"+f.Name] = true })
	return flags
}

// TestRunbookFlagsExistInHelp: every flag the runbook passes to intercom is a
// flag intercom really has, and every real flag is documented, so the runbook
// cannot drift from cmd/intercom.
func TestRunbookFlagsExistInHelp(t *testing.T) {
	doc := runbook(t)
	real := realFlags(t)

	used := map[string]bool{}
	for _, line := range strings.Split(doc, "\n") {
		m := invocation.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		for _, f := range flagToken.FindAllString(m[1], -1) {
			used[f] = true
			if !real[f] {
				t.Errorf("runbook line %q uses %s, which is not an intercom flag", strings.TrimSpace(line), f)
			}
		}
	}
	for _, want := range []string{"--workspace", "--config", "--port"} {
		if !used[want] {
			t.Errorf("the runbook's start command does not show %s", want)
		}
	}
	for f := range real {
		if f == "--help" {
			continue
		}
		if !strings.Contains(doc, "`"+f+"`") {
			t.Errorf("the runbook's flag table does not document %s", f)
		}
	}
}

// TestRunbookDocumentsEveryRemoteKey: each [remote] key (read from the config
// struct's toml tags) is described in the runbook.
func TestRunbookDocumentsEveryRemoteKey(t *testing.T) {
	doc := runbook(t)
	rt := reflect.TypeOf(config.RemoteConfig{})
	if rt.NumField() == 0 {
		t.Fatal("config.RemoteConfig has no fields")
	}
	for i := 0; i < rt.NumField(); i++ {
		key := rt.Field(i).Tag.Get("toml")
		if key == "" {
			t.Fatalf("field %s has no toml tag", rt.Field(i).Name)
		}
		if !strings.Contains(doc, "| `"+key+"` |") {
			t.Errorf("the runbook's [remote] table does not describe %q", key)
		}
	}
}

// TestRunbookStatesCredentialHandling pins the statements the acceptance
// criteria require: shown once on the local TUI, never in config or logs, and
// the accepted subprotocol logging residual.
func TestRunbookStatesCredentialHandling(t *testing.T) {
	doc := runbook(t)
	for _, want := range []string{
		"shown once, on the local TUI only",
		"Ctrl+R",
		"Sec-WebSocket-Protocol",
		"Accepted residual: subprotocol logging",
		"--allow-anonymous false",
		"`Authorization: Bearer <credential>`",
		"Bearer credentials can be replayed within their lifetime",
		"unverified against a real devtunnel build",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("the runbook no longer says %q", want)
		}
	}
}
