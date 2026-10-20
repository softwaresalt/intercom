package transport

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// TestCredentialComparisonIsConstantTimeBySource is a source guard: the
// credential digests in credential.go are compared only with
// subtle.ConstantTimeCompare, never with bytes.Equal, ==, != or reflect.
func TestCredentialComparisonIsConstantTimeBySource(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "credential.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, imp := range file.Imports {
		p, _ := strconv.Unquote(imp.Path.Value)
		if p == "bytes" || p == "reflect" {
			t.Errorf("credential.go imports %q; digest comparison must use crypto/subtle only", p)
		}
	}
	digestish := func(e ast.Expr) bool {
		hit := false
		ast.Inspect(e, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if x.Name == "digest" || x.Name == "want" {
					hit = true
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "digest" {
					hit = true
				}
			}
			return !hit
		})
		return hit
	}
	constantTimeCalls := 0
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok {
					switch {
					case id.Name == "subtle" && sel.Sel.Name == "ConstantTimeCompare":
						constantTimeCalls++
					case id.Name == "bytes" && (sel.Sel.Name == "Equal" || sel.Sel.Name == "Compare"):
						t.Errorf("%s: bytes.%s used in credential.go", fset.Position(x.Pos()), sel.Sel.Name)
					case id.Name == "reflect":
						t.Errorf("%s: reflect used in credential.go", fset.Position(x.Pos()))
					}
				}
			}
		case *ast.BinaryExpr:
			if (x.Op == token.EQL || x.Op == token.NEQ) && (digestish(x.X) || digestish(x.Y)) {
				t.Errorf("%s: digest compared with %s; use subtle.ConstantTimeCompare", fset.Position(x.Pos()), x.Op)
			}
		}
		return true
	})
	if constantTimeCalls == 0 {
		t.Fatal("credential.go no longer calls subtle.ConstantTimeCompare")
	}
}
