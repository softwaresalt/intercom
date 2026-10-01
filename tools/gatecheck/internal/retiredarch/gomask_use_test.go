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

// scanGoMaskFlow reports whether file f declares scanGo and, inside it, the
// result of gomask.MaskGoNonCode reaches the scan loop: the call's result is
// assigned to a variable v, and scanGo ranges over pysem.SplitLines(v). A
// package-level count of calls cannot prove this, because scanGo could switch
// to a local clone while a dead helper keeps a canonical call.
func scanGoMaskFlow(f *ast.File) (declared, flows bool) {
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != "scanGo" || fn.Body == nil {
			continue
		}
		declared = true
		maskedVars := map[string]bool{}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
				return true
			}
			if id, ok := as.Lhs[0].(*ast.Ident); ok && id.Name != "_" && isSelectorCall(as.Rhs[0], "gomask", "MaskGoNonCode") {
				maskedVars[id.Name] = true
			}
			return true
		})
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			rs, ok := n.(*ast.RangeStmt)
			if !ok || !isSelectorCall(rs.X, "pysem", "SplitLines") {
				return true
			}
			if arg, ok := rs.X.(*ast.CallExpr).Args[0].(*ast.Ident); ok && maskedVars[arg.Name] {
				flows = true
			}
			return true
		})
	}
	return declared, flows
}

// isSelectorCall reports whether e is a call to pkg.name with one argument.
func isSelectorCall(e ast.Expr, pkg, name string) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg && sel.Sel.Name == name
}

// TestScanGoConsumesCanonicalMask pins the canonical masker to scanGo's own
// masking branch, the data-flow half of the retired identity assertion.
func TestScanGoConsumesCanonicalMask(t *testing.T) {
	src, err := os.ReadFile("scango.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), "scango.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared, flows := scanGoMaskFlow(f)
	if !declared {
		t.Fatal("scango.go declares no scanGo")
	}
	if !flows {
		t.Error("scanGo does not range over pysem.SplitLines of a gomask.MaskGoNonCode result")
	}
}

// TestScanGoMaskFlow_Mutations proves scanGoMaskFlow is red when scanGo stops
// consuming the canonical mask result, and green on the production shape.
func TestScanGoMaskFlow_Mutations(t *testing.T) {
	const head = "package p\nfunc scanGo(text string, mask bool) {\n\tmasked := text\n"
	const loop = "\tfor range pysem.SplitLines(masked) {\n\t}\n}\n"
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"production shape", head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n" + loop, true},
		{"local clone in scanGo, canonical call in dead helper",
			head + "\tif mask {\n\t\tmasked = clone(text)\n\t}\n" + loop +
				"func dead(s string) { _ = gomask.MaskGoNonCode(s) }\n", false},
		{"result discarded", head + "\tif mask {\n\t\t_ = gomask.MaskGoNonCode(text)\n\t}\n" + loop, false},
		{"loop scans the unmasked text",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n" +
				"\tfor range pysem.SplitLines(text) {\n\t}\n}\n", false},
		{"other package masker", head + "\tif mask {\n\t\tmasked = other.MaskGoNonCode(text)\n\t}\n" + loop, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := parser.ParseFile(token.NewFileSet(), "x.go", tc.src, 0)
			if err != nil {
				t.Fatal(err)
			}
			declared, flows := scanGoMaskFlow(f)
			if !declared {
				t.Fatal("scanGo not found")
			}
			if flows != tc.want {
				t.Errorf("scanGoMaskFlow flows = %v, want %v", flows, tc.want)
			}
		})
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
