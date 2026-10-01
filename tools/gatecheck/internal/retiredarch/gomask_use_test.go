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
		ast.Inspect(f, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.FuncDecl:
				name := strings.ToLower(v.Name.Name)
				if strings.Contains(name, "mask") && strings.Contains(name, "noncode") {
					t.Errorf("%s: local masker %s shadows gomask.MaskGoNonCode", p, v.Name.Name)
				}
			case *ast.SelectorExpr:
				if id, ok := v.X.(*ast.Ident); ok && id.Name == "gomask" && v.Sel.Name == "MaskGoNonCode" {
					calls++
				}
			}
			return true
		})
	}
	if calls == 0 {
		t.Error("no call to gomask.MaskGoNonCode in package retiredarch")
	}
}
