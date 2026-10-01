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

// gomaskImportPath is the canonical masker package. Selector calls count only
// when their qualifier is bound to this import path, so an unrelated package
// imported under the name gomask cannot satisfy the guards.
const gomaskImportPath = "github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"

// gomaskName returns the file-local name bound to gomaskImportPath, or "" when
// the file does not import it under a usable name or a local declaration
// anywhere in the file reuses that name (which could shadow the import).
func gomaskName(f *ast.File) string {
	name := ""
	for _, imp := range f.Imports {
		if strings.Trim(imp.Path.Value, "`\"") != gomaskImportPath {
			continue
		}
		name = "gomask"
		if imp.Name != nil {
			name = imp.Name.Name
		}
	}
	if name == "" || name == "_" || name == "." || declaresName(f, name) {
		return ""
	}
	return name
}

// declaresName reports whether any declaration in f (function, type, const,
// var, parameter, result, receiver, := or range :=) introduces name.
func declaresName(f *ast.File, name string) bool {
	found := false
	hit := func(id *ast.Ident) {
		if id != nil && id.Name == name {
			found = true
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			hit(v.Name)
		case *ast.TypeSpec:
			hit(v.Name)
		case *ast.ValueSpec:
			for _, id := range v.Names {
				hit(id)
			}
		case *ast.Field:
			for _, id := range v.Names {
				hit(id)
			}
		case *ast.AssignStmt:
			if v.Tok == token.DEFINE {
				for _, l := range v.Lhs {
					if id, ok := l.(*ast.Ident); ok {
						hit(id)
					}
				}
			}
		case *ast.RangeStmt:
			if v.Tok == token.DEFINE {
				for _, e := range []ast.Expr{v.Key, v.Value} {
					if id, ok := e.(*ast.Ident); ok {
						hit(id)
					}
				}
			}
		}
		return true
	})
	return found
}

// maskerUse reports, for one parsed file, how many times gomask.MaskGoNonCode
// is actually called and the names of any locally declared maskers. Only
// *ast.CallExpr nodes whose Fun is the selector count: a bare reference such
// as `var _ = gomask.MaskGoNonCode` is not a call. The selector qualifier must
// be the name bound to gomaskImportPath (see gomaskName).
func maskerUse(f *ast.File) (calls int, locals []string) {
	pkg := gomaskName(f)
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.FuncDecl:
			name := strings.ToLower(v.Name.Name)
			if strings.Contains(name, "mask") && strings.Contains(name, "noncode") {
				locals = append(locals, v.Name.Name)
			}
		case *ast.CallExpr:
			if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
				if id, ok := sel.X.(*ast.Ident); ok && pkg != "" && id.Name == pkg && sel.Sel.Name == "MaskGoNonCode" {
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
// result of gomask.MaskGoNonCode reaches the scan loop: scanGo ranges over
// pysem.SplitLines(v) in a top-level statement, and the last write to v
// positioned before that range statement is a single assignment of a
// gomask.MaskGoNonCode result (qualifier bound to gomaskImportPath). A
// package-level count of calls cannot prove this, because scanGo could switch
// to a local clone while a dead helper keeps a canonical call.
//
// "Last write by source position" is a conservative stand-in for reaching-
// definition analysis, made sound by also pinning the canonical write to the
// production shape: it must be a top-level statement of scanGo, or a direct
// statement of a top-level `if p { ... }` with no init and no else whose
// condition is a scanGo parameter that scanGo never writes. A write anywhere
// else (a nested block, a loop, `if false`, a constant condition, or an
// uncalled func literal) can never be canonical, and any later write to v on
// any branch (a noncanonical reassignment, a multi-assign, a var
// redeclaration, or v's address being taken) makes the guard red, so it
// fails closed rather than open. A write to v inside any func literal is also
// red wherever the literal sits, since a closure declared before the canonical
// write can run after it.
func scanGoMaskFlow(f *ast.File) (declared, flows bool) {
	pkg := gomaskName(f)
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != "scanGo" || fn.Body == nil {
			continue
		}
		declared = true
		if pkg == "" {
			continue
		}
		allowed := canonicalSites(fn)
		for _, st := range fn.Body.List {
			rs, ok := st.(*ast.RangeStmt)
			if !ok || !isSelectorCall(rs.X, "pysem", "SplitLines") {
				continue
			}
			if arg, ok := rs.X.(*ast.CallExpr).Args[0].(*ast.Ident); ok && lastWriteIsCanonical(fn.Body, arg.Name, rs.Pos(), pkg, allowed) {
				flows = true
			}
		}
	}
	return declared, flows
}

// canonicalSites returns the statements of fn where a canonical mask write is
// structurally guaranteed to execute when reached: the top-level statements of
// fn's body, and the direct statements of a top-level IfStmt with no Init, no
// Else, and a condition that is a bare parameter of fn never written in fn.
func canonicalSites(fn *ast.FuncDecl) map[ast.Stmt]bool {
	params := map[string]bool{}
	for _, fld := range fn.Type.Params.List {
		for _, id := range fld.Names {
			params[id.Name] = true
		}
	}
	sites := map[ast.Stmt]bool{}
	for _, st := range fn.Body.List {
		sites[st] = true
		is, ok := st.(*ast.IfStmt)
		if !ok || is.Init != nil || is.Else != nil {
			continue
		}
		cond, ok := is.Cond.(*ast.Ident)
		if !ok || !params[cond.Name] || len(writePositions(fn.Body, cond.Name)) > 0 {
			continue
		}
		for _, inner := range is.Body.List {
			sites[inner] = true
		}
	}
	return sites
}

// writePositions returns the positions of every write to name in body: an
// assignment LHS, a range-clause assignment (`for k, v = range`), a var
// declaration, an inc/dec, or the address being taken.
func writePositions(body *ast.BlockStmt, name string) []token.Pos {
	var pos []token.Pos
	is := func(e ast.Expr) bool {
		id, ok := e.(*ast.Ident)
		return ok && id.Name == name
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			for _, l := range s.Lhs {
				if is(l) {
					pos = append(pos, s.Pos())
				}
			}
		case *ast.IncDecStmt:
			if is(s.X) {
				pos = append(pos, s.Pos())
			}
		case *ast.RangeStmt:
			if s.Tok == token.ASSIGN && ((s.Key != nil && is(s.Key)) || (s.Value != nil && is(s.Value))) {
				pos = append(pos, s.Pos())
			}
		case *ast.ValueSpec:
			for _, id := range s.Names {
				if id.Name == name {
					pos = append(pos, s.Pos())
				}
			}
		case *ast.UnaryExpr:
			if s.Op == token.AND && is(s.X) {
				pos = append(pos, s.Pos())
			}
		}
		return true
	})
	return pos
}

// closureWrites reports whether any func literal in body writes name. Source
// position says nothing about when a closure runs, so a captured write must
// fail the guard closed wherever the literal is declared.
func closureWrites(body *ast.BlockStmt, name string) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if fl, ok := n.(*ast.FuncLit); ok && len(writePositions(fl.Body, name)) > 0 {
			found = true
		}
		return !found
	})
	return found
}

// lastWriteIsCanonical reports whether, among all writes to name in body that
// start before limit, the last one is `name = <pkg>.MaskGoNonCode(x)` (or :=)
// located at one of the allowed canonical sites. Any write to name inside a
// func literal makes it false, because execution order is not source order.
func lastWriteIsCanonical(body *ast.BlockStmt, name string, limit token.Pos, pkg string, allowed map[ast.Stmt]bool) bool {
	if closureWrites(body, name) {
		return false
	}
	var last token.Pos
	canonical := false
	record := func(pos token.Pos, isCanonical bool) {
		if pos < limit && pos > last {
			last, canonical = pos, isCanonical
		}
	}
	for _, p := range writePositions(body, name) {
		record(p, false)
	}
	ast.Inspect(body, func(n ast.Node) bool {
		s, ok := n.(*ast.AssignStmt)
		if !ok || !allowed[s] || len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return true
		}
		if id, ok := s.Lhs[0].(*ast.Ident); ok && id.Name == name && s.Pos() == last && isSelectorCall(s.Rhs[0], pkg, "MaskGoNonCode") {
			canonical = true
		}
		return true
	})
	return canonical
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
	const imp = "import \"" + gomaskImportPath + "\"\n"
	const head = "package p\n" + imp + "func scanGo(text string, mask bool) {\n\tmasked := text\n"
	const loop = "\tfor range pysem.SplitLines(masked) {\n\t}\n}\n"
	const canon = "\t\tmasked = gomask.MaskGoNonCode(text)\n"
	cases := []struct {
		name string
		src  string
		want bool
	}{
		{"production shape", head + "\tif mask {\n" + canon + "\t}\n" + loop, true},
		{"unconditional top-level canonical write", head + "\tmasked = gomask.MaskGoNonCode(text)\n" + loop, true},
		{"canonical import under an alias",
			"package p\nimport gm \"" + gomaskImportPath + "\"\nfunc scanGo(text string, mask bool) {\n\tmasked := text\n" +
				"\tif mask {\n\t\tmasked = gm.MaskGoNonCode(text)\n\t}\n" + loop, true},
		{"local clone in scanGo, canonical call in dead helper",
			head + "\tif mask {\n\t\tmasked = clone(text)\n\t}\n" + loop +
				"func dead(s string) { _ = gomask.MaskGoNonCode(s) }\n", false},
		{"result discarded", head + "\tif mask {\n\t\t_ = gomask.MaskGoNonCode(text)\n\t}\n" + loop, false},
		{"loop scans the unmasked text",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n" +
				"\tfor range pysem.SplitLines(text) {\n\t}\n}\n", false},
		{"other package masker", head + "\tif mask {\n\t\tmasked = other.MaskGoNonCode(text)\n\t}\n" + loop, false},
		{"canonical result overwritten by clone",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t\tmasked = clone(text)\n\t}\n" + loop, false},
		{"canonical result overwritten after the branch",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n\tmasked = text\n" + loop, false},
		{"canonical result overwritten by multi-assign",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n\tmasked, _ = text, 0\n" + loop, false},
		{"canonical result overwritten by range value assign",
			head + "\tif mask {\n" + canon + "\t}\n\tfor _, masked = range []string{text} {\n\t}\n" + loop, false},
		{"canonical result overwritten by range key assign",
			head + "\tif mask {\n" + canon + "\t}\n\tfor masked = range map[string]int{text: 0} {\n\t}\n" + loop, false},
		{"canonical result shadowed by var",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n\tvar masked = text\n" + loop, false},
		{"address taken after canonical write",
			head + "\tif mask {\n\t\tmasked = gomask.MaskGoNonCode(text)\n\t}\n\tclobber(&masked)\n" + loop, false},
		{"foreign package imported as gomask",
			"package p\nimport gomask \"example.com/other\"\nfunc scanGo(text string, mask bool) {\n\tmasked := text\n" +
				"\tif mask {\n" + canon + "\t}\n" + loop, false},
		{"canonical package not imported", strings.Replace(head, imp, "", 1) + "\tif mask {\n" + canon + "\t}\n" + loop, false},
		{"local gomask shadows the import", head + "\tgomask := other\n\tif mask {\n" + canon + "\t}\n" + loop, false},
		{"canonical write under if false", head + "\tif false {\n" + canon + "\t}\n" + loop, false},
		{"canonical write under constant condition", head + "\tconst on = true\n\tif on {\n" + canon + "\t}\n" + loop, false},
		{"condition parameter reassigned", head + "\tmask = false\n\tif mask {\n" + canon + "\t}\n" + loop, false},
		{"canonical write under if with init", head + "\tif _ = 0; mask {\n" + canon + "\t}\n" + loop, false},
		{"canonical write under if with else", head + "\tif mask {\n" + canon + "\t} else {\n\t}\n" + loop, false},
		{"canonical write in uncalled func literal", head + "\t_ = func() {\n" + canon + "\t}\n" + loop, false},
		{"closure declared before canonical write, called after",
			head + "\tclobber := func() { masked = text }\n\tif mask {\n" + canon + "\t}\n\tclobber()\n" + loop, false},
		{"closure compound-assigns a captured mask",
			head + "\tclobber := func() { masked += text }\n\tif mask {\n" + canon + "\t}\n\tclobber()\n" + loop, false},
		{"canonical write in nested loop", head + "\tfor range []int{} {\n" + canon + "\t}\n" + loop, false},
		{"canonical write in nested block under if", head + "\tif mask {\n\t\t{\n" + canon + "\t\t}\n\t}\n" + loop, false},
		{"scan loop not top-level",
			head + "\tif mask {\n" + canon + "\t}\n\t{\n\t\tfor range pysem.SplitLines(masked) {\n\t\t}\n\t}\n}\n", false},
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
	const imp = "import \"" + gomaskImportPath + "\"\n"
	cases := []struct {
		name       string
		src        string
		wantCalls  int
		wantLocals int
	}{
		{"real call", "package p\n" + imp + "func f(b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 1, 0},
		{"aliased canonical import", "package p\nimport gm \"" + gomaskImportPath + "\"\nfunc f(b []byte) { _ = gm.MaskGoNonCode(b) }\n", 1, 0},
		{"bare reference", "package p\n" + imp + "var _ = gomask.MaskGoNonCode\n", 0, 0},
		{"reference passed as value", "package p\n" + imp + "func f() { g(gomask.MaskGoNonCode) }\n", 0, 0},
		{"local masker", "package p\nfunc maskGoNonCode(b []byte) []byte { return b }\n", 0, 1},
		{"other package selector", "package p\n" + imp + "func f(b []byte) { _ = other.MaskGoNonCode(b) }\n", 0, 0},
		{"foreign package imported as gomask", "package p\nimport gomask \"example.com/other\"\nfunc f(b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 0, 0},
		{"canonical package not imported", "package p\nfunc f(b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 0, 0},
		{"blank import", "package p\nimport _ \"" + gomaskImportPath + "\"\nfunc f(b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 0, 0},
		{"local var shadows the import", "package p\n" + imp + "func f(b []byte) { gomask := other; _ = gomask.MaskGoNonCode(b) }\n", 0, 0},
		{"parameter shadows the import", "package p\n" + imp + "func f(gomask T, b []byte) { _ = gomask.MaskGoNonCode(b) }\n", 0, 0},
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
