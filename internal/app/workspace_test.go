package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
)

func loadCfg(t *testing.T, root, tables string) *config.Config {
	t.Helper()
	cfg, _, err := config.Load(writeConfig(t, root, "", tables))
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return cfg
}

func canonical(t *testing.T, dir string) string {
	t.Helper()
	r, err := pathsafe.NewRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	return r.Path()
}

func wantViolation(t *testing.T, err error, what string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: accepted; want a refusal", what)
	}
	var ae *apperr.Error
	if !errors.As(err, &ae) || (ae.Kind() != apperr.KindPathViolation && ae.Kind() != apperr.KindConfig) {
		t.Errorf("%s: error %v is not a path-violation/config apperr", what, err)
	}
}

func TestResolveWorkspace(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "proj", "inner")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir() // a sibling temp dir: outside every authorized root
	cfg := loadCfg(t, root, "")

	t.Run("empty selects the default root", func(t *testing.T) {
		got, err := resolveWorkspace(cfg, "")
		if err != nil || got != canonical(t, root) {
			t.Fatalf("resolveWorkspace(\"\") = %q, %v; want %q", got, err, canonical(t, root))
		}
	})
	t.Run("the root itself", func(t *testing.T) {
		got, err := resolveWorkspace(cfg, root)
		if err != nil || got != canonical(t, root) {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("a directory inside the root", func(t *testing.T) {
		got, err := resolveWorkspace(cfg, sub)
		if err != nil || got != canonical(t, sub) {
			t.Fatalf("got %q, %v; want %q", got, err, canonical(t, sub))
		}
	})
	t.Run("a directory outside every root", func(t *testing.T) {
		_, err := resolveWorkspace(cfg, other)
		wantViolation(t, err, "outside directory")
	})
	t.Run("dot-dot escape", func(t *testing.T) {
		_, err := resolveWorkspace(cfg, filepath.Join(sub, "..", "..", "..", filepath.Base(other)))
		wantViolation(t, err, "dot-dot escape")
	})
	t.Run("sibling sharing the root's name prefix", func(t *testing.T) {
		sib := root + "-evil"
		if err := os.MkdirAll(sib, 0o750); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(sib) })
		_, err := resolveWorkspace(cfg, sib)
		wantViolation(t, err, "prefix sibling")
	})
	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := resolveWorkspace(cfg, filepath.Join(root, "absent"))
		wantViolation(t, err, "missing directory")
	})
	t.Run("a file, not a directory", func(t *testing.T) {
		file := filepath.Join(root, "file.txt")
		if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := resolveWorkspace(cfg, file)
		wantViolation(t, err, "file")
	})
	t.Run("symlink inside the root pointing outside", func(t *testing.T) {
		link := filepath.Join(root, "escape")
		if err := os.Symlink(other, link); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
		_, err := resolveWorkspace(cfg, link)
		wantViolation(t, err, "symlink escape")
	})
}

func TestResolveWorkspaceHonoursMappedRoots(t *testing.T) {
	root, mapped := t.TempDir(), t.TempDir()
	cfg := loadCfg(t, root, fmt.Sprintf("[[workspace]]\nworkspace_id = \"w1\"\npath = '%s'\n", mapped))
	got, err := resolveWorkspace(cfg, mapped)
	if err != nil || got != canonical(t, mapped) {
		t.Fatalf("mapped workspace root refused: %q, %v", got, err)
	}
}
