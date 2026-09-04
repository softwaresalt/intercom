package config

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/BurntSushi/toml"
)

// exampleConfigPath locates config.toml.example, which lives at the
// repository root — two directories above this package.
func exampleConfigPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "config.toml.example")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config.toml.example not found at %s: %v", path, err)
	}
	return path
}

func TestExampleConfigLoadsWithZeroUnknownKeys(t *testing.T) {
	cfg, report, err := Load(exampleConfigPath(t))
	if err != nil {
		t.Fatalf("Load(config.toml.example) returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load(config.toml.example) returned nil *Config")
	}
	if len(report.UnknownKeys) != 0 {
		t.Errorf("Report.UnknownKeys = %v, want empty (every example key must match a schema field)", report.UnknownKeys)
	}
}

func TestExampleConfigMatchesDefaultsOutsideDocumentedExemptions(t *testing.T) {
	exemptions := []string{
		"default_workspace_root",
		"commands",
		"workspace",
	}

	got, _, err := Load(exampleConfigPath(t))
	if err != nil {
		t.Fatalf("Load(config.toml.example) returned unexpected error: %v", err)
	}
	want := Default()

	if got.MaxConcurrentSessions != want.MaxConcurrentSessions {
		t.Errorf("config.toml.example max_concurrent_sessions = %v, want default %v", got.MaxConcurrentSessions, want.MaxConcurrentSessions)
	}
	if got.Copilot != want.Copilot {
		t.Errorf("config.toml.example [copilot] = %+v, want default %+v", got.Copilot, want.Copilot)
	}
	if got.HTTPPort != want.HTTPPort {
		t.Errorf("config.toml.example http_port = %v, want default %v", got.HTTPPort, want.HTTPPort)
	}
	if got.RetentionDays != want.RetentionDays {
		t.Errorf("config.toml.example retention_days = %v, want default %v", got.RetentionDays, want.RetentionDays)
	}
	if got.OperatorDetailLevel != want.OperatorDetailLevel {
		t.Errorf("config.toml.example operator_detail_level = %v, want default %v", got.OperatorDetailLevel, want.OperatorDetailLevel)
	}
	if got.Timeouts != want.Timeouts {
		t.Errorf("config.toml.example [timeouts] = %+v, want default %+v", got.Timeouts, want.Timeouts)
	}
	if got.Stall != want.Stall {
		t.Errorf("config.toml.example [stall] = %+v, want default %+v", got.Stall, want.Stall)
	}
	if got.Database != want.Database {
		t.Errorf("config.toml.example [database] = %+v, want default %+v", got.Database, want.Database)
	}

	_ = exemptions
}

func TestExampleConfigCoversEveryStructField(t *testing.T) {
	expected := collectTOMLKeyPaths(reflect.TypeOf(Config{}), "")
	sort.Strings(expected)

	data, err := os.ReadFile(exampleConfigPath(t))
	if err != nil {
		t.Fatalf("failed to read config.toml.example: %v", err)
	}

	var probe Config
	md, err := toml.Decode(string(data), &probe)
	if err != nil {
		t.Fatalf("toml.Decode(config.toml.example) returned unexpected error: %v", err)
	}

	present := make(map[string]struct{})
	for _, k := range md.Keys() {
		present[k.String()] = struct{}{}
	}

	var missing []string
	for _, path := range expected {
		if _, ok := present[path]; !ok {
			missing = append(missing, path)
		}
	}
	if len(missing) != 0 {
		t.Errorf("config.toml.example is missing key paths present in the Config struct hierarchy: %v", missing)
	}
}

// collectTOMLKeyPaths walks t (a struct type) and returns every dotted
// TOML key path implied by its "toml" struct tags, recursing one level
// into nested struct fields and into the element type of a slice of
// structs (without an index — toml.MetaData.Keys() reports array-of-
// tables paths without an index too, per unit D1/E1 verification against
// the library's actual behavior).
func collectTOMLKeyPaths(t reflect.Type, prefix string) []string {
	var paths []string
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("toml")
		if tag == "" || tag == "-" {
			continue
		}

		path := tag
		if prefix != "" {
			path = prefix + "." + tag
		}
		paths = append(paths, path)

		ft := field.Type
		switch ft.Kind() {
		case reflect.Struct:
			paths = append(paths, collectTOMLKeyPaths(ft, path)...)
		case reflect.Slice:
			if ft.Elem().Kind() == reflect.Struct {
				paths = append(paths, collectTOMLKeyPaths(ft.Elem(), path)...)
			}
		}
	}
	return paths
}
