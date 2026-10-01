package retiredarch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// maskerUse reports, for one parsed file, how many times gomask.MaskGoNonCode
// is actually called and the names of any locally declared maskers. Only
// *ast.CallExpr nodes whose Fun is the selector count: a bare reference such
// as `var _ = gomask.MaskGoNonCode` is not a call.
func maskerUse(f *ast.File) (calls int, locals []string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			name := strings.ToLower(v.Name.Name)
			if strings.Contains(name, "mask") && strings.Contains(name, "noncode") {
				locals = append(locals, v.Name.Name)
			}
		case *ast.CallExpr:
			if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && id.Name == "gomask" && sel.Sel.Name == "MaskGoNonCode" {
					calls++
				}
			}
		}
		return true
	})
	return calls, locals
}

// TestUsesSharedCanonicalMasker is the Go counterpart of the retired Python
// test_uses_shared_canonical_masker. The compiler only proves that gomask is
// imported; this test proves the package calls gomask.MaskGoNonCode and
// declares no local masker of its own that could drift from the canonical one.
func TestUsesSharedCanonicalMasker(t *testing.T) {
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	calls := 0
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		c, locals := maskerUse(f)
		calls += c
		for _, l := range locals {
			t.Errorf("%s: local masker %s shadows gomask.MaskGoNonCode", p, l)
		}
	}
	if calls == 0 {
		t.Error("no call to gomask.MaskGoNonCode in package retiredarch")
	}
}

// TestMaskerUse_Mutations proves maskerUse is red on the regressions it
// guards against and green on a genuine call.
func TestMaskerUse_Mutations(t *testing.T) {
	cases := []struct {
		name       string
		src        string
		wantCalls  int
		wantLocals int
	}{
		{"real call", "package p\nfunc f(b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 1, 0},
		{"bare reference", "package p\nvar _ = gomask.MaskGoNonCode\n", 0, 0},
		{"reference passed as value", "package p\nfunc f() { g(gomask.MaskGoNonCode) }\n", 0, 0},
		{"local masker", "package p\nfunc maskGoNonCode(b []byte) []byte { return b }\n", 0, 1},
		{"other package selector", "package p\nfunc f(b []byte) { _ = other.MaskGoNonCode(b) }\n", 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "x.go", tc.src, 0)
			if err != nil {
				t.Fatal(err)
			}
			calls, locals := maskerUse(f)
			if calls != tc.wantCalls || len(locals) != tc.wantLocals {
				t.Errorf("maskerUse = (%d, %q), want (%d calls, %d locals)", calls, locals, tc.wantCalls, tc.wantLocals)
			}
		})
	}
}
