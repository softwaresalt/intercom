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

// exampleValueAgreementExemptions enumerates the top-level toml keys that
// are legitimately illustrative rather than default-equal: an empty root
// path only validates against the package test working directory, and the
// two collection fields exist purely to demonstrate shape. Every other
// top-level field must match Default() exactly, checked generically below
// so a future field addition is covered without a corresponding hand-written
// comparison (U-D2 acceptance criterion: the assertion must fail if
// database.path reverts to agent-intercom.db).
func exampleValueAgreementExemptions() map[string]struct{} {
	return map[string]struct{}{
		"default_workspace_root": {},
		"commands":               {},
		"workspace":              {},
	}
}

func TestExampleConfigMatchesDefaultsOutsideDocumentedExemptions(t *testing.T) {
	exemptions := exampleValueAgreementExemptions()

	got, _, err := Load(exampleConfigPath(t))
	if err != nil {
		t.Fatalf("Load(config.toml.example) returned unexpected error: %v", err)
	}
	want := Default()

	gotVal := reflect.ValueOf(got).Elem()
	wantVal := reflect.ValueOf(want).Elem()
	typ := gotVal.Type()

	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get("toml")
		if tag == "" || tag == "-" {
			continue
		}
		if _, exempt := exemptions[tag]; exempt {
			continue
		}

		gotField := gotVal.Field(i).Interface()
		wantField := wantVal.Field(i).Interface()
		if !reflect.DeepEqual(gotField, wantField) {
			t.Errorf("config.toml.example %s = %+v, want default %+v", tag, gotField, wantField)
		}
	}
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
